// Package factory is ORM's source-agnostic co-generation core: a Source builds a
// coreir.Model, a Target renders it in a chosen language, and a Registry wires
// the two together so one binary can drive proto→DB targets and
// GraphQL→client targets from one config.
//
// It sits ABOVE protokit (which is a proto→DB-schema engine welded to protoc):
// the proto Source wraps protokit, so protokit's registry stays untouched
// underneath while this layer adds the second source and the CLI surface.
package factory

import (
	"sort"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"

	"github.com/the-protobuf-project/orm/plugin/factory/coreir"
)

// Ctx threads generation context through the pipeline. Plugin is set in plugin
// (protoc) mode — where buf/protoc hand ORM a CodeGeneratorRequest — and nil in
// CLI mode; targets that need it (the proto/DB targets) check and error when it
// is absent.
type Ctx struct {
	Plugin *protogen.Plugin
}

// Source builds a coreir.Model from some input (proto descriptors via protokit,
// a GraphQL endpoint/schema via introspection, …).
type Source interface {
	Name() string
	Build(Ctx) (*coreir.Model, error)
}

// Target renders a coreir.Model into output files for one language.
type Target interface {
	Name() string
	// Languages lists the languages this target can emit; the factory validates a
	// requested language against it. Go-only today.
	Languages() []string
	Generate(ctx Ctx, m *coreir.Model, lang string) error
}

// Registry holds the sources and targets a binary ships, each keyed by Name.
type Registry struct {
	Sources map[string]Source
	Targets map[string]Target
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{Sources: map[string]Source{}, Targets: map[string]Target{}}
}

// AddSource registers s under s.Name().
func (r *Registry) AddSource(s Source) { r.Sources[s.Name()] = s }

// AddTarget registers t under t.Name().
func (r *Registry) AddTarget(t Target) { r.Targets[t.Name()] = t }

// TargetNames returns the registered target names, sorted and comma-joined, for
// error messages.
func (r *Registry) TargetNames() string {
	names := make([]string, 0, len(r.Targets))
	for k := range r.Targets {
		names = append(names, k)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}
