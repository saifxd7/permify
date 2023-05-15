package engines

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/factories"
	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
	"github.com/Permify/permify/pkg/tuple"
)

var _ = Describe("expand-engine", func() {
	// DRIVE SAMPLE
	driveSchema := `
	entity user {}

	entity organization {
		relation admin @user
	}

	entity folder {
		relation org @organization
		relation creator @user
		relation collaborator @user

		permission read = collaborator
		permission update = collaborator
		permission delete = creator or org.admin
	}

	entity doc {
		relation org @organization
		relation parent @folder
		relation owner @user @folder#creator

		permission read = (owner or parent.collaborator) or org.admin
		permission update = owner and not org.admin
		permission delete = owner or not update
		permission view = owner and not read
	}
	`

	Context("Drive Sample: Expand", func() {
		It("Drive Sample: Case 1", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			// SCHEMA

			conf, err := newSchema(driveSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			// RELATIONSHIPS

			type expand struct {
				entity     string
				assertions map[string]*base.Expand
			}

			tests := struct {
				relationships []string
				expands       []expand
			}{
				relationships: []string{
					"doc:1#owner@user:2",
					"doc:1#parent@folder:1#...",
					"folder:1#collaborator@user:1",
					"folder:1#collaborator@user:3",
					"doc:1#org@organization:1#...",
					"organization:1#admin@user:1",
					"folder:2#creator@user:89",
					"doc:1#owner@folder:2#creator",
				},
				expands: []expand{
					{
						entity: "doc:1",
						assertions: map[string]*base.Expand{
							"read": {
								Target: &base.EntityAndRelation{
									Entity: &base.Entity{
										Type: "doc",
										Id:   "1",
									},
									Relation: "read",
								},
								Node: &base.Expand_Expand{
									Expand: &base.ExpandTreeNode{
										Operation: base.ExpandTreeNode_OPERATION_UNION,
										Children: []*base.Expand{
											{
												Target: &base.EntityAndRelation{
													Entity: &base.Entity{
														Type: "doc",
														Id:   "1",
													},
													Relation: "read",
												},
												Node: &base.Expand_Expand{
													Expand: &base.ExpandTreeNode{
														Operation: base.ExpandTreeNode_OPERATION_UNION,
														Children: []*base.Expand{
															{
																Target: &base.EntityAndRelation{
																	Entity: &base.Entity{
																		Type: "doc",
																		Id:   "1",
																	},
																	Relation: "owner",
																},
																Node: &base.Expand_Expand{
																	Expand: &base.ExpandTreeNode{
																		Operation: base.ExpandTreeNode_OPERATION_UNION,
																		Children: []*base.Expand{
																			{
																				Target: &base.EntityAndRelation{
																					Entity: &base.Entity{
																						Type: "folder",
																						Id:   "2",
																					},
																					Relation: "creator",
																				},
																				Node: &base.Expand_Leaf{
																					Leaf: &base.Subjects{
																						Subjects: []*base.Subject{
																							{
																								Type: "user",
																								Id:   "89",
																							},
																						},
																					},
																				},
																			},
																			{
																				Target: &base.EntityAndRelation{
																					Entity: &base.Entity{
																						Type: "doc",
																						Id:   "1",
																					},
																					Relation: "owner",
																				},
																				Node: &base.Expand_Leaf{
																					Leaf: &base.Subjects{
																						Subjects: []*base.Subject{
																							{
																								Type: "user",
																								Id:   "2",
																							},
																							{
																								Type:     "folder",
																								Id:       "2",
																								Relation: "creator",
																							},
																						},
																					},
																				},
																			},
																		},
																	},
																},
															},
															{
																Target: &base.EntityAndRelation{
																	Entity: &base.Entity{
																		Type: "doc",
																		Id:   "1",
																	},
																	Relation: "read",
																},
																Node: &base.Expand_Expand{
																	Expand: &base.ExpandTreeNode{
																		Operation: base.ExpandTreeNode_OPERATION_UNION,
																		Children: []*base.Expand{
																			{
																				Target: &base.EntityAndRelation{
																					Entity: &base.Entity{
																						Type: "folder",
																						Id:   "1",
																					},
																					Relation: "collaborator",
																				},
																				Node: &base.Expand_Leaf{
																					Leaf: &base.Subjects{
																						Subjects: []*base.Subject{
																							{
																								Type: "user",
																								Id:   "1",
																							},
																							{
																								Type: "user",
																								Id:   "3",
																							},
																						},
																					},
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
											{
												Target: &base.EntityAndRelation{
													Entity: &base.Entity{
														Type: "doc",
														Id:   "1",
													},
													Relation: "read",
												},
												Node: &base.Expand_Expand{
													Expand: &base.ExpandTreeNode{
														Operation: base.ExpandTreeNode_OPERATION_UNION,
														Children: []*base.Expand{
															{
																Target: &base.EntityAndRelation{
																	Entity: &base.Entity{
																		Type: "organization",
																		Id:   "1",
																	},
																	Relation: "admin",
																},
																Node: &base.Expand_Leaf{
																	Leaf: &base.Subjects{
																		Subjects: []*base.Subject{
																			{
																				Type: "user",
																				Id:   "1",
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						entity: "doc:1",
						assertions: map[string]*base.Expand{
							"delete": {
								Target: &base.EntityAndRelation{
									Entity: &base.Entity{
										Type: "doc",
										Id:   "1",
									},
									Relation: "delete",
								},
								Node: &base.Expand_Expand{
									Expand: &base.ExpandTreeNode{
										Operation: base.ExpandTreeNode_OPERATION_UNION,
										Children: []*base.Expand{
											{
												Target: &base.EntityAndRelation{
													Entity: &base.Entity{
														Type: "doc",
														Id:   "1",
													},
													Relation: "owner",
												},
												Node: &base.Expand_Expand{
													Expand: &base.ExpandTreeNode{
														Operation: base.ExpandTreeNode_OPERATION_UNION,
														Children: []*base.Expand{
															{
																Target: &base.EntityAndRelation{
																	Entity: &base.Entity{
																		Type: "folder",
																		Id:   "2",
																	},
																	Relation: "creator",
																},
																Node: &base.Expand_Leaf{
																	Leaf: &base.Subjects{
																		Subjects: []*base.Subject{
																			{
																				Type: "user",
																				Id:   "89",
																			},
																		},
																	},
																},
															},
															{
																Target: &base.EntityAndRelation{
																	Entity: &base.Entity{
																		Type: "doc",
																		Id:   "1",
																	},
																	Relation: "owner",
																},
																Node: &base.Expand_Leaf{
																	Leaf: &base.Subjects{
																		Subjects: []*base.Subject{
																			{
																				Type: "user",
																				Id:   "2",
																			},
																			{
																				Type:     "folder",
																				Id:       "2",
																				Relation: "creator",
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
											{
												Target: &base.EntityAndRelation{
													Entity: &base.Entity{
														Type: "doc",
														Id:   "1",
													},
													Relation: "update",
												},
												Node: &base.Expand_Expand{
													Expand: &base.ExpandTreeNode{
														Operation: base.ExpandTreeNode_OPERATION_UNION,
														Children: []*base.Expand{
															{
																Target: &base.EntityAndRelation{
																	Entity: &base.Entity{
																		Type: "doc",
																		Id:   "1",
																	},
																	Relation: "owner",
																},
																Node: &base.Expand_Expand{
																	Expand: &base.ExpandTreeNode{
																		Operation: base.ExpandTreeNode_OPERATION_UNION,
																		Children: []*base.Expand{
																			{
																				Target: &base.EntityAndRelation{
																					Entity: &base.Entity{
																						Type: "folder",
																						Id:   "2",
																					},
																					Relation: "creator",
																				},
																				Exclusion: true,
																				Node: &base.Expand_Leaf{
																					Leaf: &base.Subjects{
																						Subjects: []*base.Subject{
																							{
																								Type: "user",
																								Id:   "89",
																							},
																						},
																					},
																				},
																			},
																			{
																				Target: &base.EntityAndRelation{
																					Entity: &base.Entity{
																						Type: "doc",
																						Id:   "1",
																					},
																					Relation: "owner",
																				},
																				Exclusion: true,
																				Node: &base.Expand_Leaf{
																					Leaf: &base.Subjects{
																						Subjects: []*base.Subject{
																							{
																								Type: "user",
																								Id:   "2",
																							},
																							{
																								Type:     "folder",
																								Id:       "2",
																								Relation: "creator",
																							},
																						},
																					},
																				},
																			},
																		},
																	},
																},
															},
															{
																Target: &base.EntityAndRelation{
																	Entity: &base.Entity{
																		Type: "doc",
																		Id:   "1",
																	},
																	Relation: "update",
																},
																Node: &base.Expand_Expand{
																	Expand: &base.ExpandTreeNode{
																		Operation: base.ExpandTreeNode_OPERATION_UNION,
																		Children: []*base.Expand{
																			{
																				Target: &base.EntityAndRelation{
																					Entity: &base.Entity{
																						Type: "organization",
																						Id:   "1",
																					},
																					Relation: "admin",
																				},
																				Node: &base.Expand_Leaf{
																					Leaf: &base.Subjects{
																						Subjects: []*base.Subject{
																							{
																								Type: "user",
																								Id:   "1",
																							},
																						},
																					},
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						entity: "doc:1",
						assertions: map[string]*base.Expand{
							"view": {
								Target: &base.EntityAndRelation{
									Entity: &base.Entity{
										Type: "doc",
										Id:   "1",
									},
									Relation: "view",
								},
								Node: &base.Expand_Expand{
									Expand: &base.ExpandTreeNode{
										Operation: base.ExpandTreeNode_OPERATION_INTERSECTION,
										Children: []*base.Expand{
											{
												Target: &base.EntityAndRelation{
													Entity: &base.Entity{
														Type: "doc",
														Id:   "1",
													},
													Relation: "owner",
												},
												Node: &base.Expand_Expand{
													Expand: &base.ExpandTreeNode{
														Operation: base.ExpandTreeNode_OPERATION_UNION,
														Children: []*base.Expand{
															{
																Target: &base.EntityAndRelation{
																	Entity: &base.Entity{
																		Type: "folder",
																		Id:   "2",
																	},
																	Relation: "creator",
																},
																Node: &base.Expand_Leaf{
																	Leaf: &base.Subjects{
																		Subjects: []*base.Subject{
																			{
																				Type: "user",
																				Id:   "89",
																			},
																		},
																	},
																},
															},
															{
																Target: &base.EntityAndRelation{
																	Entity: &base.Entity{
																		Type: "doc",
																		Id:   "1",
																	},
																	Relation: "owner",
																},
																Node: &base.Expand_Leaf{
																	Leaf: &base.Subjects{
																		Subjects: []*base.Subject{
																			{
																				Type: "user",
																				Id:   "2",
																			},
																			{
																				Type:     "folder",
																				Id:       "2",
																				Relation: "creator",
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
											{
												Target: &base.EntityAndRelation{
													Entity: &base.Entity{
														Type: "doc",
														Id:   "1",
													},
													Relation: "read",
												},
												Node: &base.Expand_Expand{
													Expand: &base.ExpandTreeNode{
														Operation: base.ExpandTreeNode_OPERATION_INTERSECTION,
														Children: []*base.Expand{
															{
																Target: &base.EntityAndRelation{
																	Entity: &base.Entity{
																		Type: "doc",
																		Id:   "1",
																	},
																	Relation: "read",
																},
																Node: &base.Expand_Expand{
																	Expand: &base.ExpandTreeNode{
																		Operation: base.ExpandTreeNode_OPERATION_INTERSECTION,
																		Children: []*base.Expand{
																			{
																				Target: &base.EntityAndRelation{
																					Entity: &base.Entity{
																						Type: "doc",
																						Id:   "1",
																					},
																					Relation: "owner",
																				},
																				Node: &base.Expand_Expand{
																					Expand: &base.ExpandTreeNode{
																						Operation: base.ExpandTreeNode_OPERATION_UNION,
																						Children: []*base.Expand{
																							{
																								Target: &base.EntityAndRelation{
																									Entity: &base.Entity{
																										Type: "folder",
																										Id:   "2",
																									},
																									Relation: "creator",
																								},
																								Exclusion: true,
																								Node: &base.Expand_Leaf{
																									Leaf: &base.Subjects{
																										Subjects: []*base.Subject{
																											{
																												Type: "user",
																												Id:   "89",
																											},
																										},
																									},
																								},
																							},
																							{
																								Target: &base.EntityAndRelation{
																									Entity: &base.Entity{
																										Type: "doc",
																										Id:   "1",
																									},
																									Relation: "owner",
																								},
																								Exclusion: true,
																								Node: &base.Expand_Leaf{
																									Leaf: &base.Subjects{
																										Subjects: []*base.Subject{
																											{
																												Type: "user",
																												Id:   "2",
																											},
																											{
																												Type:     "folder",
																												Id:       "2",
																												Relation: "creator",
																											},
																										},
																									},
																								},
																							},
																						},
																					},
																				},
																			},
																			{
																				Target: &base.EntityAndRelation{
																					Entity: &base.Entity{
																						Type: "doc",
																						Id:   "1",
																					},
																					Relation: "read",
																				},
																				Node: &base.Expand_Expand{
																					Expand: &base.ExpandTreeNode{
																						Operation: base.ExpandTreeNode_OPERATION_UNION,
																						Children: []*base.Expand{
																							{
																								Target: &base.EntityAndRelation{
																									Entity: &base.Entity{
																										Type: "folder",
																										Id:   "1",
																									},
																									Relation: "collaborator",
																								},
																								Exclusion: true,
																								Node: &base.Expand_Leaf{
																									Leaf: &base.Subjects{
																										Subjects: []*base.Subject{
																											{
																												Type: "user",
																												Id:   "1",
																											},
																											{
																												Type: "user",
																												Id:   "3",
																											},
																										},
																									},
																								},
																							},
																						},
																					},
																				},
																			},
																		},
																	},
																},
															},
															{
																Target: &base.EntityAndRelation{
																	Entity: &base.Entity{
																		Type: "doc",
																		Id:   "1",
																	},
																	Relation: "read",
																},
																Node: &base.Expand_Expand{
																	Expand: &base.ExpandTreeNode{
																		Operation: base.ExpandTreeNode_OPERATION_UNION,
																		Children: []*base.Expand{
																			{
																				Target: &base.EntityAndRelation{
																					Entity: &base.Entity{
																						Type: "organization",
																						Id:   "1",
																					},
																					Relation: "admin",
																				},
																				Exclusion: true,
																				Node: &base.Expand_Leaf{
																					Leaf: &base.Subjects{
																						Subjects: []*base.Subject{
																							{
																								Type: "user",
																								Id:   "1",
																							},
																						},
																					},
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			relationshipReader := factories.RelationshipReaderFactory(db, logger.New("debug"))
			relationshipWriter := factories.RelationshipWriterFactory(db, logger.New("debug"))

			expandEngine := NewExpandEngine(schemaReader, relationshipReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				relationshipReader,
				nil,
				expandEngine,
				nil,
				nil,
			)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = relationshipWriter.WriteRelationships(context.Background(), "t1", database.NewTupleCollection(tuples...))
			Expect(err).ShouldNot(HaveOccurred())

			for _, expand := range tests.expands {
				entity, err := tuple.E(expand.entity)
				Expect(err).ShouldNot(HaveOccurred())

				for permission, res := range expand.assertions {
					var response *base.PermissionExpandResponse
					response, err = invoker.Expand(context.Background(), &base.PermissionExpandRequest{
						TenantId:   "t1",
						Entity:     entity,
						Permission: permission,
						Metadata: &base.PermissionExpandRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.Tree).Should(Equal(res))
				}
			}
		})
	})
})
