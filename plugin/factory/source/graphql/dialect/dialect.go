// Package dialect abstracts the engine-specific conventions of a GraphQL backend
// so the client generator is not welded to one server. The classifiers that read
// a schema (what marks a bool_exp, a comparison, a mutation verb, a mutation
// result, a sort input) all resolve their magic strings through a Dialect, so a
// new engine (Hasura today; others later) is a new Dialect value plus a registry
// entry — no changes to the IR builder or renderer.
package dialect

import (
	"fmt"
	"sort"
	"strings"
)

// Verb maps one CRUD mutation convention onto a friendly Go method name.
// OpPrefix is the lowercase operation-name prefix on a root mutation field
// (e.g. "insert" in insertVenues); NamePrefix is the PascalCase form used when
// re-homing procedure-style mutations onto their table family (e.g. "Insert");
// Friendly is the emitted method verb (e.g. "Create").
type Verb struct {
	OpPrefix   string
	NamePrefix string
	Friendly   string
}

// Dialect is the set of conventions one GraphQL engine follows. All methods are
// pure (no schema access); the IR builder and renderer consult them instead of
// hardcoding operator/field/verb names.
type Dialect interface {
	// Name is the dialect's registry key (e.g. "hasura").
	Name() string

	// AuthHeader is the HTTP header an admin secret is sent under, or "" if the
	// engine has no such convention.
	AuthHeader() string

	// Combinators are the boolean-expression combinator field names (and, or, not).
	Combinators() (and, or, not string)

	// EqOperators are the comparison-input operand fields, canonical-equality first
	// (e.g. {"_eq","_in"}). The first entry doubles as the marker that an input is a
	// scalar comparison rather than a row filter.
	EqOperators() []string

	// SetOperands are the field names that carry an update column's new value inside
	// a per-column set input (e.g. {"set","_set"}).
	SetOperands() []string

	// MutationVerbs list the CRUD verb conventions, in match-precedence order.
	MutationVerbs() []Verb

	// ByIdSuffix is the root-field suffix marking a by-primary-key lookup (e.g. "ById").
	ByIdSuffix() string

	// ReturningField is the mutation-response field holding the affected rows.
	ReturningField() string

	// AffectedRowsFields are the mutation-response affected-count field names.
	AffectedRowsFields() []string

	// AggregateSuffixes are the object-name suffixes marking an aggregate wrapper
	// over a row type (e.g. {"AggExp","Aggregate"}).
	AggregateSuffixes() []string

	// PreCheckArgs are the optimistic-concurrency precondition argument names.
	PreCheckArgs() []string

	// DefaultScalars maps the engine's known GraphQL scalar names to Go types.
	DefaultScalars() map[string]string
}

// registry holds the dialects a build ships, keyed by Name.
var registry = map[string]Dialect{}

// Register adds d to the registry. Called from dialect init() functions.
func Register(d Dialect) { registry[d.Name()] = d }

// Get returns the dialect named name, or an error listing the registered names.
func Get(name string) (Dialect, error) {
	d, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown graphql dialect %q — registered: %s", name, Registered())
	}
	return d, nil
}

// Registered returns the sorted, comma-joined registered dialect names.
func Registered() string {
	names := make([]string, 0, len(registry))
	for k := range registry {
		names = append(names, k)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}
