// Package generator composes the database backends protoc-gen-orm ships —
// gorm, sql, and prisma — into a target registry for the protokit run harness.
// The IR frontend and shared infrastructure now live in the protokit module;
// this package is just orm's database-target composition point.
package generator

import (
	"github.com/the-protobuf-project/protokit/schema"
	"github.com/the-protobuf-project/orm/plugin/generator/gorm"
	"github.com/the-protobuf-project/orm/plugin/generator/prisma"
	sqlgen "github.com/the-protobuf-project/orm/plugin/generator/sql"
)

// Targets is the database-backend registry, keyed by the value users write in
// buf.gen.yaml opt: [target=<key>].
func Targets() map[string]schema.Target {
	return map[string]schema.Target{
		"gorm":   &gorm.Generator{},
		"sql":    &sqlgen.Generator{},
		"prisma": &prisma.Generator{},
	}
}
