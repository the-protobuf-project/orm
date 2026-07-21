package gorm

// telemetry.go reads the orm.v1 telemetry options off the IR's Source
// descriptors at render time — the same pattern filterspec.go uses for query
// options — and resolves each table's effective instrumentation: the tree-wide
// telemetry opt gates everything, (orm.v1.telemetry) tunes one table, and
// (orm.v1.telemetry_field) projects a field into span attributes. Metrics stay
// DB-scoped by design (table/op/status only) — domain-level metrics belong to
// the application, not the ORM.

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/the-protobuf-project/orm/plugin/pb/ormpbv1"
	"github.com/the-protobuf-project/protokit/header"
	"github.com/the-protobuf-project/protokit/naming"
	"github.com/the-protobuf-project/protokit/schema"
)

// opentelementryModule is the import path of the first-party observability SDK
// every telemetry-generated file goes through. The plugin itself never imports
// it — only generated consumers do.
const opentelementryModule = "github.com/the-protobuf-project/opentelementry/opentelementry-go"

// ormtelemetryPkg is the package name and output directory of the generated
// SDK adapter at <go_module>/ormtelemetry — the only generated package that
// imports the SDK.
const ormtelemetryPkg = "ormtelemetry"

// tableTelemetry resolves one table's effective instrumentation: enabled only
// when the tree-wide telemetry opt is on and the table doesn't opt out via
// (orm.v1.telemetry).enabled=false; metrics narrows further per table; the
// span prefix defaults to "<schema>.<Model>".
func tableTelemetry(db *schema.Database, s *schema.Schema, t *schema.Table) (enabled, metrics bool, spanPrefix string) {
	if !dbTelemetry(db) {
		return false, false, ""
	}
	o := telemetryOpts(t.Source)
	if o.Enabled != nil && !o.GetEnabled() {
		return false, false, ""
	}
	metrics = dbTelemetryMetrics(db)
	if o.Metrics != nil {
		metrics = o.GetMetrics()
	}
	spanPrefix = o.GetSpanPrefix()
	if spanPrefix == "" {
		spanPrefix = s.Name + "." + t.LocalName
	}
	return true, metrics, spanPrefix
}

// telemetryTag renders a column's opentelementry struct tag — a
// "trace:<name>" directive the SDK lifts into a span attribute on traced
// writes. Empty when the table isn't instrumented or the field isn't labeled.
func telemetryTag(enabled bool, t *schema.Table, col *schema.Column) string {
	if !enabled {
		return ""
	}
	o := telemetryFieldOpts(col.Source)
	if !o.GetLabel() {
		return ""
	}
	return `opentelementry:"trace:` + telemetryAttrName(t, col, o.GetLabelName()) + `"`
}

// telemetryAttrName is the span-attribute name for a field: the explicit
// override, or "<model_snake>.<column>" (e.g. "book.genre").
func telemetryAttrName(t *schema.Table, col *schema.Column, override string) string {
	if override != "" {
		return override
	}
	return naming.SnakeCase(t.LocalName) + "." + col.Name
}

// ormtelemetryView assembles the data for the once-per-tree ormtelemetry
// package: the SDK adapter (with stores) and the gorm query plugin.
func ormtelemetryView(db *schema.Database) map[string]any {
	return map[string]any{
		"Header": header.Render("//", header.Info{
			PluginVersion: db.PluginVersion,
			ProtocVersion: db.ProtocVersion,
			Database:      db.Name,
			SchemaLabel:   "package",
			Schema:        ormtelemetryPkg,
			Notes:         []string{"First-party opentelementry adapter: the stores' gormx.Telemetry and the SQL-level gorm plugin."},
		}),
		"Package":              ormtelemetryPkg,
		"Stores":               dbStores(db),
		"Metrics":              dbTelemetryMetrics(db),
		"Logs":                 dbTelemetryLogs(db),
		"OpentelementryImport": opentelementryModule,
		"GormxImport":          dbGoModule(db) + "/" + gormxPkg,
	}
}

// --- option accessors (safe empty value when the extension or descriptor is absent) ---

func telemetryOpts(d protoreflect.MessageDescriptor) *ormpbv1.TelemetryOptions {
	if d == nil || !proto.HasExtension(d.Options(), ormpbv1.E_Telemetry) {
		return &ormpbv1.TelemetryOptions{}
	}
	return proto.GetExtension(d.Options(), ormpbv1.E_Telemetry).(*ormpbv1.TelemetryOptions)
}

func telemetryFieldOpts(d protoreflect.FieldDescriptor) *ormpbv1.TelemetryFieldOptions {
	if d == nil || !proto.HasExtension(d.Options(), ormpbv1.E_TelemetryField) {
		return &ormpbv1.TelemetryFieldOptions{}
	}
	return proto.GetExtension(d.Options(), ormpbv1.E_TelemetryField).(*ormpbv1.TelemetryFieldOptions)
}
