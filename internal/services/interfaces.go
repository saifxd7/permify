package services

import (
	"context"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/repositories/filters"
	"github.com/Permify/permify/pkg/tuple"
)

// IPermissionService -
type IPermissionService interface {
	Check(ctx context.Context, subject tuple.Subject, action string, entity tuple.Entity, version string, d int32) (response commands.CheckResponse, err error)
	Expand(ctx context.Context, entity tuple.Entity, action string, version string) (response commands.ExpandResponse, err error)
}

// IRelationshipService -
type IRelationshipService interface {
	ReadRelationships(ctx context.Context, filter filters.RelationTupleFilter) ([]tuple.Tuple, error)
	WriteRelationship(ctx context.Context, entities tuple.Tuple, version string) error
	DeleteRelationship(ctx context.Context, entities tuple.Tuple) error
}

// ISchemaService -
type ISchemaService interface {
	Lookup(ctx context.Context, entityType string, relationNames []string, version string) (response commands.SchemaLookupResponse, err error)
}
