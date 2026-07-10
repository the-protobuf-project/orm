{{.Header}}

package {{.Package}}

import (
	"{{.RepoxImport}}"
)

{{- range .Resources}}

// Format{{.Model}}Name builds "{{.Pattern}}" from bare ids.
func Format{{.Model}}Name({{.VarList}} string) string {
	return {{.FormatExpr}}
}

// Parse{{.Model}}Name splits a "{{.Pattern}}" resource name into its bare ids,
// rejecting other shapes with repox.ErrInvalidArgument.
func Parse{{.Model}}Name(name string) ({{.VarList}} string, err error) {
	ids, err := repox.SplitName(name, {{.CollectionsExpr}})
	if err != nil {
		return
	}
	{{- range $i, $v := .Vars}}
	{{$v}} = ids[{{$i}}]
	{{- end}}
	return
}
{{- end}}
