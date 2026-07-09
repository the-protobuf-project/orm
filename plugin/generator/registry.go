// Package generator composes the database backends protoc-gen-orm ships —
// gorm, sql, and prisma — into a target registry for the protokit run harness.
// The IR frontend and shared infrastructure now live in the protokit module;
// this package is just orm's database-target composition point.
package generator

import (
	"github.com/the-protobuf-project/orm/plugin/factory"
	"github.com/the-protobuf-project/orm/plugin/factory/target/dbtarget"
	"github.com/the-protobuf-project/orm/plugin/generator/gorm"
	"github.com/the-protobuf-project/orm/plugin/generator/prisma"
	"github.com/the-protobuf-project/orm/plugin/generator/sql"
	"github.com/the-protobuf-project/protokit/schema"
)

// Targets is the database-backend registry, keyed by the value users write in
// buf.gen.yaml opt: [target=<key>]. It is the raw protokit-target view used by
// the golden test harness; the plugin binary drives these through the factory
// (see FactoryDBTargets).
func Targets() map[string]schema.Target {
	return map[string]schema.Target{
		"gorm":   &gorm.Generator{},
		"sql":    &sql.Generator{},
		"prisma": &prisma.Generator{},
	}
}

// FactoryDBTargets adapts every database backend to a factory.Target, so the
// factory registry can drive them alongside non-proto targets (e.g. the GraphQL
// client) from one dispatch path.
func FactoryDBTargets() []factory.Target {
	raw := Targets()
	out := make([]factory.Target, 0, len(raw))
	for _, t := range raw {
		out = append(out, dbtarget.Wrap(t))
	}
	return out
}
