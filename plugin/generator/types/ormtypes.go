// Package types is orm's projection of the neutral schema.FieldType onto the
// canonical PostgreSQL type the gorm/sql/prisma targets render from. It is the
// db-specific half of the type system that used to live in protokit; protokit now
// carries only the neutral FieldType, and orm's own type override
// (orm.v1.column.type/max_length/precision) is read here off the column's Source
// descriptor rather than being stored back on the shared IR.
package types

import (
	"fmt"

	"github.com/the-protobuf-project/orm/plugin/pb/ormpbv1"
	"github.com/the-protobuf-project/protokit/schema"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// sqlForType maps a neutral FieldType to its canonical PostgreSQL type, matching
// what protokit's PostgresType produced before the split. Unsigned 32/64 widen a
// step so the full range fits. TypeEnum has no SQL type (enum columns carry
// schema.Column.Enum instead).
var sqlForType = map[schema.FieldType]string{
	schema.TypeString:    "VARCHAR(255)",
	schema.TypeBool:      "BOOLEAN",
	schema.TypeInt32:     "INTEGER",
	schema.TypeUint32:    "BIGINT",
	schema.TypeInt64:     "BIGINT",
	schema.TypeUint64:    "NUMERIC(20,0)",
	schema.TypeFloat:     "REAL",
	schema.TypeDouble:    "DOUBLE PRECISION",
	schema.TypeBytes:     "BYTEA",
	schema.TypeTimestamp: "TIMESTAMPTZ",
	schema.TypeDuration:  "INTERVAL",
	schema.TypeDate:      "DATE",
	schema.TypeTimeOfDay: "TIME",
	schema.TypeDecimal:   "NUMERIC",
	schema.TypeLatLng:    "POINT",
	schema.TypeInterval:  "TSTZRANGE",
	schema.TypeText:      "TEXT",
	schema.TypeJSON:      "JSONB",
	schema.TypeULID:      "CHAR(26)",
	schema.TypeUUID:      "UUID",
}

// SQL returns the canonical PostgreSQL type for a neutral FieldType, appending the
// array suffix for a repeated field. JSONB stays a single document (one JSON value
// already represents the whole collection).
func SQL(t schema.FieldType, list bool) string {
	base := sqlForType[t]
	if list && base != "" && t != schema.TypeJSON {
		return base + "[]"
	}
	return base
}

// SQLForColumn returns the effective PostgreSQL type of a column: an explicit
// orm.v1.column type/max_length/precision override (read off the column's Source
// descriptor) wins; otherwise the neutral FieldType — proto-classified, or set by
// protokit's synthesis and foreign-key alignment — projects to a PostgreSQL type.
func SQLForColumn(col *schema.Column) string {
	if t := overrideType(col.Source); t != "" {
		return t
	}
	return SQL(col.Type, col.List)
}

// overrideType returns the SQL type an orm.v1.column type/max_length/precision
// override pins on a field, or "" when the field has no such override (or is a
// synthesized column with no descriptor).
func overrideType(d protoreflect.FieldDescriptor) string {
	o := columnOpts(d)
	switch {
	case o.GetType() != "":
		return o.GetType()
	case o.GetMaxLength() > 0:
		return fmt.Sprintf("VARCHAR(%d)", o.GetMaxLength())
	case o.GetPrecision() > 0:
		return fmt.Sprintf("NUMERIC(%d,%d)", o.GetPrecision(), o.GetScale())
	}
	return ""
}

// columnOpts reads the orm.v1.column extension off a field descriptor, returning a
// safe empty value when the descriptor is nil or carries no annotation.
func columnOpts(d protoreflect.FieldDescriptor) *ormpbv1.ColumnOptions {
	if d == nil || !proto.HasExtension(d.Options(), ormpbv1.E_Column) {
		return &ormpbv1.ColumnOptions{}
	}
	return proto.GetExtension(d.Options(), ormpbv1.E_Column).(*ormpbv1.ColumnOptions)
}
