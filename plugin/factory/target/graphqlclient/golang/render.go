// Package golang renders a Go client from the IR using an interface/handler
// architecture grouped by domain. Each resource is its own package exposing
// Query/Mutation/Subscription interfaces backed by unexported handlers, plus a natural
// calling surface: a predicate DSL (field handles + And/Or/Not), single-object
// CreateInput/UpdateInput structs, and chained request builders for optional arguments.
// A shared schema package holds models and a shared enums package holds enums; domain and
// root packages aggregate the handlers so callers write
// s.Query.<Domain>.<Resource>.<Method>(...).
package golang

import (
	"fmt"
	"sort"
	"strings"

	"github.com/the-protobuf-project/orm/plugin/factory/source/graphql/dialect"
	"github.com/the-protobuf-project/orm/plugin/factory/source/graphql/ir"
	"github.com/the-protobuf-project/orm/plugin/factory/target/graphqlclient/naming"
	"github.com/the-protobuf-project/orm/plugin/factory/target/graphqlclient/selection"
	"github.com/the-protobuf-project/orm/plugin/factory/target/graphqlclient/typemap"
)

// Qualifiers for referencing generated types from each writing context. Models inline
// their relations, so they only reference enums; resource-package code references models
// ("schema.") and enums ("enums.") plus the runtime graphql helpers.
// Generated packages carry a "ql" suffix on their name (protobuf-style: the import path is
// the folder, the package identifier is <folder>ql), so qualifiers reference schemaql./enumsql.
// The runtime helper package (graphql.) is not generated and keeps its name.
var (
	qModels   = typemap.Qualifier{Enums: "enumsql."}
	qResource = typemap.Qualifier{Models: "schemaql.", Enums: "enumsql."}
)

// op pairs an operation with its de-duplicated exported method name. The list query also
// produces a second op (FindOne) that reuses the list operation but returns the first match
// as a single pointer; ListName carries the sibling List method whose request type it shares.
type op struct {
	Op       *ir.Operation
	Name     string
	FindOne  bool   // render a "first match" wrapper returning *Row instead of []Row
	ListName string // sibling List method (and request type) a FindOne op reuses
}

// requestName is the operation whose <Name>Request type this op uses. A FindOne op shares
// the sibling List request rather than declaring its own identical one.
func (o op) requestName() string {
	if o.FindOne {
		return o.ListName
	}
	return o.Name
}

// renderer turns IR elements into Go source fragments.
type renderer struct {
	schema    *ir.Schema
	mapper    *typemap.Mapper
	selection *selection.Renderer
	dialect   dialect.Dialect
}

// comparisonMarker is the field whose presence marks an input as a scalar
// comparison (rather than a row filter) — the dialect's canonical equality op.
func (r *renderer) comparisonMarker() string { return r.dialect.EqOperators()[0] }

// verbByFriendly returns the dialect verb whose friendly name matches (e.g.
// "Create"), reporting false when the dialect defines no such verb.
func (r *renderer) verbByFriendly(friendly string) (dialect.Verb, bool) {
	for _, v := range r.dialect.MutationVerbs() {
		if v.Friendly == friendly {
			return v, true
		}
	}
	return dialect.Verb{}, false
}

// enum renders a Go enum: a named string type plus a typed constant per value.
func (r *renderer) enum(e *ir.Enum) string {
	var b strings.Builder
	fmt.Fprintf(&b, "// %s is the %s enum.\ntype %s string\n\n", e.Name, e.Name, e.Name)
	if len(e.Values) > 0 {
		b.WriteString("const (\n")
		for _, v := range e.Values {
			fmt.Fprintf(&b, "\t%s%s %s = %q\n", e.Name, naming.Export(v), e.Name, v)
		}
		b.WriteString(")\n")
	}
	return b.String()
}

// model renders an object's model struct, with relations inlined to max depth.
func (r *renderer) model(o *ir.Object) string {
	body := r.selection.ModelBody(o)
	var b strings.Builder
	fmt.Fprintf(&b, "// %s is the %s model.\ntype %s struct {\n%s}\n", o.Name, o.Name, o.Name, body)
	return b.String()
}

// sortedKeys returns the keys of m in deterministic order.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// sortedFields returns input fields sorted by name for deterministic output.
func sortedFields(fields []ir.Field) []ir.Field {
	out := append([]ir.Field(nil), fields...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
