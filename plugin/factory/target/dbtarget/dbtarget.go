// Package dbtarget adapts protokit's proto-bound schema.Target (gorm, sql,
// prisma) to the factory's source-agnostic factory.Target. One generic wrapper
// serves all three DB backends: it reads the Model's DBSchema facet and delegates
// to the underlying renderer with the protoc plugin context, so the existing
// generators keep producing byte-identical output while joining the factory.
package dbtarget

import (
	"fmt"

	"github.com/the-protobuf-project/protokit/schema"

	"github.com/the-protobuf-project/orm/plugin/factory"
	"github.com/the-protobuf-project/orm/plugin/factory/coreir"
)

// Target wraps one protokit schema.Target.
type Target struct{ inner schema.Target }

// Wrap adapts inner to a factory.Target.
func Wrap(inner schema.Target) *Target { return &Target{inner: inner} }

// Name returns the wrapped target's name (e.g. "gorm").
func (t *Target) Name() string { return t.inner.Name() }

// Languages: the DB targets emit Go (structs/DDL/Prisma project files) today.
func (t *Target) Languages() []string { return []string{"go"} }

// Generate renders the DBSchema facet through the wrapped protokit target. It
// needs the protoc plugin context, so it is only usable in plugin mode.
func (t *Target) Generate(ctx factory.Ctx, m *coreir.Model, _ string) error {
	if ctx.Plugin == nil {
		return fmt.Errorf("target %q requires a protoc plugin context (proto sources run under buf/protoc, not the CLI)", t.inner.Name())
	}
	if m == nil || m.DBSchema == nil {
		return fmt.Errorf("target %q: model has no proto/DB schema (wrong source?)", t.inner.Name())
	}
	return t.inner.Generate(ctx.Plugin, m.DBSchema)
}
