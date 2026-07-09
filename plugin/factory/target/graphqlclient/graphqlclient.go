// Package graphqlclient is the factory Target that renders a typed Go GraphQL
// client from the Model's GraphQL facet. It adapts the golang renderer to the
// factory.Target interface, carrying its output configuration (module, package,
// runtime) since — unlike the proto/DB targets — it writes real files to a
// directory rather than through a protoc plugin response.
package graphqlclient

import (
	"fmt"
	"path"
	"strings"

	"github.com/the-protobuf-project/orm/plugin/factory"
	"github.com/the-protobuf-project/orm/plugin/factory/coreir"
	"github.com/the-protobuf-project/orm/plugin/factory/source/graphql/dialect"
	"github.com/the-protobuf-project/orm/plugin/factory/target/graphqlclient/golang"
)

// defaultRuntimeModule is the import path of the Go runtime facade generated
// clients depend on; the sibling predicate-DSL package is derived from it.
const defaultRuntimeModule = "github.com/the-protobuf-project/runtime-go/network/runtime"

// Config is the output configuration for one client generation.
type Config struct {
	OutDir        string            // parent dir; client lands in <OutDir>/<Package>/
	Package       string            // root package name; default: base(GoModule)+"ql"
	GoModule      string            // import path of the generated root package (required)
	RuntimeModule string            // runtime facade import path; default defaultRuntimeModule
	MaxDepth      int               // relation inlining depth (default 1)
	Scalars       map[string]string // GraphQL scalar -> Go type overrides
	Dialect       dialect.Dialect   // engine conventions; default dialect.Default()

	// Sink, when set, receives each generated file (path relative to the package
	// root, content) instead of writing to OutDir. The plugin sets it to route
	// output through the protoc response.
	Sink func(relPath string, content []byte) error
}

// Target renders the GraphQL client.
type Target struct{ cfg Config }

// New returns a graphql-client target, applying defaults.
func New(cfg Config) *Target {
	if cfg.Dialect == nil {
		cfg.Dialect = dialect.Default()
	}
	if cfg.RuntimeModule == "" {
		cfg.RuntimeModule = defaultRuntimeModule
	}
	if cfg.MaxDepth == 0 {
		cfg.MaxDepth = 1
	}
	if cfg.Package == "" && cfg.GoModule != "" {
		cfg.Package = defaultPackage(cfg.GoModule)
	}
	return &Target{cfg: cfg}
}

// defaultPackage derives the root package name from the module path: the last
// segment, suffixed with "ql" unless it already ends in it.
func defaultPackage(goModule string) string {
	p := path.Base(goModule)
	if !strings.HasSuffix(p, "ql") {
		p += "ql"
	}
	return p
}

// Name is the registry/config key.
func (t *Target) Name() string { return "graphql-client" }

// PackageName is the resolved root package name (folder the client is written to).
func (t *Target) PackageName() string { return t.cfg.Package }

// Languages: the client is emitted in Go today.
func (t *Target) Languages() []string { return []string{"go"} }

// Generate renders the client from the Model's GraphQL facet.
func (t *Target) Generate(_ factory.Ctx, m *coreir.Model, _ string) error {
	if m == nil || m.GraphQL == nil {
		return fmt.Errorf("graphql-client target: model has no GraphQL schema (wrong source?)")
	}
	if t.cfg.GoModule == "" {
		return fmt.Errorf("graphql-client target: go_module is required (import path of the generated package)")
	}
	return golang.Generate(golang.Options{
		Schema:        m.GraphQL,
		OutDir:        t.cfg.OutDir,
		Package:       t.cfg.Package,
		GoModule:      t.cfg.GoModule,
		RuntimeModule: t.cfg.RuntimeModule,
		MaxDepth:      t.cfg.MaxDepth,
		Scalars:       t.cfg.Scalars,
		Dialect:       t.cfg.Dialect,
		Sink:          t.cfg.Sink,
	})
}
