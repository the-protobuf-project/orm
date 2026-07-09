package golang

import (
	"path"
	"sort"
	"strings"
	"unicode"

	"github.com/the-protobuf-project/orm/plugin/factory/source/graphql/ir"
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
			grp := modelGroup{file: modelFile(rg.res.Name, usedFiles)}
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
	for _, name := range sortedKeys(g.opts.Schema.Enums) {
		body := g.r.enum(g.opts.Schema.Enums[name])
		if err := g.writeFile(enumsDir, typeFile(name), "enumsql", g.typeImports(body), body); err != nil {
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

// typeFile returns the file name for a generated enum: the schema type name itself,
// kept PascalCase so it contains no underscores. Underscores would let Go misread a
// trailing word as a GOOS/GOARCH build constraint (e.g. "..._windows.go").
func typeFile(name string) string {
	return name + ".go"
}

// modelFile returns the snake_case file name for one resource's model group
// ("PropertyUnits" -> "property_units.go"). A trailing word Go would read as an implicit
// build constraint (a GOOS/GOARCH, or "test") gets a "_schema" suffix, and used guards
// against two resources snaking to the same name.
func modelFile(resource string, used map[string]bool) string {
	s := snake(resource)
	if constrainedSuffix[s[strings.LastIndex(s, "_")+1:]] {
		s += "_schema"
	}
	return uniqueName(s, used) + ".go"
}

// snake converts a PascalCase type name to snake_case, keeping acronym runs together
// ("IDDocuments" -> "id_documents").
func snake(name string) string {
	rs := []rune(name)
	var b strings.Builder
	for i, r := range rs {
		if unicode.IsUpper(r) {
			prevLower := i > 0 && (unicode.IsLower(rs[i-1]) || unicode.IsDigit(rs[i-1]))
			nextLower := i+1 < len(rs) && unicode.IsLower(rs[i+1])
			if i > 0 && (prevLower || nextLower) {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// constrainedSuffix lists the trailing filename words the Go toolchain treats as implicit
// build constraints (GOOS/GOARCH values) or as a test file marker.
var constrainedSuffix = map[string]bool{
	"aix": true, "android": true, "darwin": true, "dragonfly": true, "freebsd": true,
	"hurd": true, "illumos": true, "ios": true, "js": true, "linux": true, "nacl": true,
	"netbsd": true, "openbsd": true, "plan9": true, "solaris": true, "wasip1": true,
	"windows": true, "zos": true,
	"386": true, "amd64": true, "amd64p32": true, "arm": true, "arm64": true,
	"arm64be": true, "armbe": true, "loong64": true, "mips": true, "mips64": true,
	"mips64le": true, "mips64p32": true, "mips64p32le": true, "mipsle": true, "ppc": true,
	"ppc64": true, "ppc64le": true, "riscv": true, "riscv64": true, "s390": true,
	"s390x": true, "sparc": true, "sparc64": true, "wasm": true,
	"test": true,
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
