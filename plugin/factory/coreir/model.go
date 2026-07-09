// Package coreir is the factory's shared intermediate model. A Source builds a
// Model; a Target renders from it. The Model deliberately carries the source's
// native IR as a typed facet rather than flattening everything into one lossy
// schema: the proto/DB world (tables, columns, relations, migrations) and the
// GraphQL world (objects, inputs, operations) do not share one representation,
// so each Target reads the facet it needs. Cross-cutting reuse lives at the
// factory / config / emit layer, not in a single god-IR.
package coreir

import (
	"github.com/the-protobuf-project/orm/plugin/factory/source/graphql/introspect"
	"github.com/the-protobuf-project/orm/plugin/factory/source/graphql/ir"
	"github.com/the-protobuf-project/protokit/schema"
)

// Model is the unit of work handed from a Source to a Target. Exactly one facet
// is populated, matching the Source that built it.
type Model struct {
	// DBSchema is the proto-derived database IR (set by the proto Source; read by
	// the prisma / gorm / sql targets). Nil for non-proto sources.
	DBSchema []*schema.Database

	// GraphQL holds the introspection-derived GraphQL IR (set by the graphql
	// Source; read by the graphql-client target). Nil for non-graphql sources.
	GraphQL *ir.Schema

	// RawGraphQL is the introspection payload the GraphQL IR was built from,
	// retained so `dump_schema` can persist the exact schema. Nil otherwise.
	RawGraphQL *introspect.Schema
}
