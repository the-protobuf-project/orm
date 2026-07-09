// Package naming converts GraphQL identifiers to idiomatic Go identifiers.
package naming

import "strings"

// Export turns a GraphQL field, argument, or type name into an exported Go
// identifier. Underscore-separated segments are PascalCased and joined (so "order_by"
// becomes "OrderBy" and "_and" becomes "And"); existing camelCase/PascalCase segments
// keep their internal capitalization ("displayName" -> "DisplayName").
func Export(name string) string {
	parts := strings.Split(name, "_")
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		r := []rune(p)
		r[0] = []rune(strings.ToUpper(string(r[0])))[0]
		b.WriteString(string(r))
	}
	if b.Len() == 0 {
		return "Field"
	}
	return b.String()
}
