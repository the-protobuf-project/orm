package wire

// graphqltarget.go is the only wire file that imports the GraphQL *target*
// package (see graphqlsource.go for why the two are split — bare `graphql`
// identifier in each, no import alias).

import (
	"github.com/the-protobuf-project/orm/plugin/factory/coreir"
	"github.com/the-protobuf-project/orm/plugin/factory/target/graphql"
	"github.com/the-protobuf-project/protokit/factory"
	"github.com/the-protobuf-project/protokit/graphql/dialect"
)

// NewGraphQLTarget builds the GraphQL client target with its output configuration.
// It returns the concrete *graphql.Target so callers can read PackageName; a caller
// obtains it via this constructor and uses its methods without importing the
// graphql target package (Go needs the import only to name the package).
func NewGraphQLTarget(pkg, goModule, runtimeModule string, maxDepth int, scalars map[string]string, d dialect.Dialect, sink func(relPath string, content []byte) error) *graphql.Target {
	return graphql.New(graphql.Config{
		Package:       pkg,
		GoModule:      goModule,
		RuntimeModule: runtimeModule,
		MaxDepth:      maxDepth,
		Scalars:       scalars,
		Dialect:       d,
		Sink:          sink,
	})
}

// addGraphQLTarget registers a zero-config graphql target (for validation/listing).
func addGraphQLTarget(reg *factory.Registry[*coreir.Model]) {
	reg.AddTarget(graphql.New(graphql.Config{}))
}
