// Package proto is the factory Source that reads proto descriptors. It wraps
// protokit's proto→IR build so the rest of the factory never depends on protoc
// directly: Build runs the same protokit.BuildIR the plugin has always run and
// carries the resulting databases as the Model's DBSchema facet.
package proto

import (
	"fmt"

	"github.com/the-protobuf-project/protokit"
	"github.com/the-protobuf-project/protokit/schema"

	"github.com/the-protobuf-project/protokit/factory"
	"github.com/the-protobuf-project/orm/plugin/factory/coreir"
)

// Source builds the proto/DB IR. opts and backend are fixed at construction (the
// protoc plugin context arrives per-run via factory.Ctx).
type Source struct {
	opts    protokit.Options
	backend schema.Backend
}

// New returns a proto Source driven by protokit opts and the orm backend.
func New(opts protokit.Options, backend schema.Backend) *Source {
	return &Source{opts: opts, backend: backend}
}

// Name identifies this source in the registry and config.
func (s *Source) Name() string { return "proto" }

// Build runs protokit's IR build against the plugin's CodeGeneratorRequest.
func (s *Source) Build(ctx factory.Ctx) (*coreir.Model, error) {
	if ctx.Plugin == nil {
		return nil, fmt.Errorf("proto source requires a protoc plugin context (only available in a buf/protoc run)")
	}
	dbs, err := protokit.BuildIR(ctx.Plugin, s.opts, s.backend)
	if err != nil {
		return nil, err
	}
	return &coreir.Model{DBSchema: dbs}, nil
}
