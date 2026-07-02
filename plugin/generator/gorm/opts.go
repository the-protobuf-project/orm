package gorm

// opts.go reads the gorm target's render-time knobs off the neutral IR. The orm
// backend stamps these onto db.Opts during Enrich (protokit itself never
// interprets them); the gorm target is the only consumer.

import "github.com/the-protobuf-project/protokit/schema"

// dbGoModule is the Go import path of the generated output directory, needed to
// import each per-schema models package from the migration aggregator. Empty
// when unset (no aggregator is emitted).
func dbGoModule(db *schema.Database) string { return db.Opt("go_module") }

// dbStores reports whether to emit a typed CRUD store per resource.
func dbStores(db *schema.Database) bool { return db.Opt("stores") == "true" }

// dbOTel reports whether to fold the OpenTelemetry tracing helper into the
// migration Registry.
func dbOTel(db *schema.Database) bool { return db.Opt("otel") == "true" }

// dbOTelMetrics reports whether the generated Instrument emits metrics in
// addition to spans (only meaningful when dbOTel is true).
func dbOTelMetrics(db *schema.Database) bool { return db.Opt("otel_metrics") == "true" }
