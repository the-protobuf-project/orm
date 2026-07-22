module github.com/the-protobuf-project/orm

go 1.26.4

require (
	github.com/the-protobuf-project/protokit v1.1.0
	google.golang.org/genproto/googleapis/api v0.0.0-20260630182238-925bb5da69e7
	google.golang.org/protobuf v1.36.11
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/bufbuild/protocompile v0.14.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/the-protobuf-project/opentelementry/opentelementry-go v0.0.0-00010101000000-000000000000
	github.com/vektah/gqlparser/v2 v2.5.36 // indirect
	golang.org/x/sync v0.21.0 // indirect
)

// replace points at the sibling repo checkout so plugin/factory/target/gorm can pick up
// the new telemetry.v1 Go stubs (telemetrypbv1) before opentelementry-go publishes them —
// the published pseudo-version examples/go.mod requires predates those stubs. Local-only,
// pending a real publish of the opentelementry-go changes.
replace github.com/the-protobuf-project/opentelementry/opentelementry-go => ../opentelementry/opentelementry-go

replace github.com/the-protobuf-project/runtime-go/telemetry => ../runtime-go/telemetry
