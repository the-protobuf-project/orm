package golang

// write_types.go writes the shared model (schema) and enum packages and resolves each body's non-schema imports.

import (
	"path"
	"sort"
	"strings"

	"github.com/the-protobuf-project/protokit/graphql/ir"
	"github.com/the-protobuf-project/protokit/naming"
)

// enumsDir is the shared enums package, referenced as "enums.". Models are written into a
// per-domain "<domain>/schema" package instead of one global package, and aliased back into
// the domain aggregator (see writeDomain).
const enumsDir = "enumsql"

// modelGroup is one resource's slice of a domain schema package: every model object that
// resource's operations return (the row type, its aggregate, and the mutation responses),
// written together into one snake_case file named after the resource (e.g. all
// PropertyUnits* types land in property_units.go).
type modelGroup struct {
	file    string
	objects []*ir.Object
}

// domainObjects returns, per domain, the model objects that the domain's operations return
// (rows, responses, aggregates), grouped per resource. Because every object inlines its
// relations, this reachable set is self-contained — no cross-domain references — so each
// domain's schema package compiles on its own. An object returned by several resources is
// claimed by the first (resources are already sorted), so each type is written exactly once.
func (g *generator) domainObjects() map[string][]modelGroup {
	out := map[string][]modelGroup{}
	for _, dg := range g.domains {
		claimed := map[string]bool{}
		usedFiles := map[string]bool{}
		for _, rg := range dg.reses {
			seen := map[string]*ir.Object{}
			for _, set := range [][]op{rg.queries, rg.mutations, rg.subs} {
				for _, o := range set {
					if obj, ok := g.opts.Schema.Objects[o.Op.Return.Base]; ok && !claimed[obj.Name] {
						seen[obj.Name] = obj
					}
				}
			}
			if len(seen) == 0 {
				continue
			}
			grp := modelGroup{file: naming.GoFileName(rg.res.Name, "schema", usedFiles)}
			for _, name := range sortedKeys(seen) {
				claimed[name] = true
				grp.objects = append(grp.objects, seen[name])
			}
			out[dg.name] = append(out[dg.name], grp)
		}
		sort.Slice(out[dg.name], func(i, j int) bool { return out[dg.name][i].file < out[dg.name][j].file })
	}
	return out
}

// writeTypes writes the shared enums package and, per domain, that domain's model objects
// into "<domain>/schema", one file per resource.
func (g *generator) writeTypes() error {
	enumUsed := map[string]bool{}
	for _, name := range sortedKeys(g.opts.Schema.Enums) {
		body := g.r.enum(g.opts.Schema.Enums[name])
		if err := g.writeFile(enumsDir, naming.GoFileName(name, "enum", enumUsed), "enumsql", g.typeImports(body), body); err != nil {
			return err
		}
	}
	for _, domain := range sortedKeys(g.domSchema) {
		dir := domain + "/schemaql"
		for _, grp := range g.domSchema[domain] {
			bodies := make([]string, 0, len(grp.objects))
			for _, obj := range grp.objects {
				bodies = append(bodies, g.r.model(obj))
			}
			body := strings.Join(bodies, "\n\n")
			if err := g.writeFile(dir, grp.file, "schemaql", g.typeImports(body), body); err != nil {
				return err
			}
		}
	}
	return nil
}

// typeImports returns the non-schema imports a body needs (json, the runtime graphql
// helpers, and the shared enums package). The per-domain schema import is resolved by the
// caller, which knows the domain.
func (g *generator) typeImports(body string) []string {
	var im []string
	if strings.Contains(body, "json.RawMessage") {
		im = append(im, "encoding/json")
	}
	if strings.Contains(body, "graphql.") {
		im = append(im, g.graphqlModule())
	}
	if strings.Contains(body, "enumsql.") {
		im = append(im, g.opts.GoModule+"/"+enumsDir)
	}
	return im
}

// schemaModule is the import path of a domain's model package.
func (g *generator) schemaModule(domain string) string {
	return g.opts.GoModule + "/" + domain + "/schemaql"
}

// graphqlModule is the import path of the runtime graphql helper package (scalar types,
// predicate DSL, and column helpers), a sibling of the runtime facade module.
func (g *generator) graphqlModule() string {
	return path.Dir(g.opts.RuntimeModule) + "/graphql"
}
