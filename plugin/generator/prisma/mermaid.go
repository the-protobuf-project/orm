package prisma

// mermaid.go renders the visual + per-model detail sections of a generated
// README: the Mermaid ER diagram (PK/FK columns, crow's-foot edges, external
// stub entities) and the leaf model/enum tables. Split from readme.go, which
// keeps the directory-tree walking and page assembly.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/oh-tarnished/protorm/plugin/generator/schema"
	"github.com/oh-tarnished/protorm/plugin/generator/types"
)

// mermaidER renders an erDiagram with PK/FK columns and crow's-foot edges.
// Models referenced but not in this set become external stub entities.
func mermaidER(tables []*schema.Table) string {
	if len(tables) == 0 {
		return ""
	}
	defined := map[string]bool{}
	for _, t := range tables {
		defined[t.ModelName] = true
	}

	var b strings.Builder
	b.WriteString("## Entity relationships\n\n```mermaid\nerDiagram\n    direction LR\n")

	stubs := map[string]bool{}
	sorted := append([]*schema.Table(nil), tables...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ModelName < sorted[j].ModelName })

	for _, t := range sorted {
		pk := map[string]bool{}
		if t.PKColumn != "" {
			pk[t.PKColumn] = true
		}
		fk := map[string]bool{}
		for _, f := range t.ForeignKeys {
			fk[f.Column] = true
			if !defined[f.ReferencedModel] {
				stubs[f.ReferencedModel] = true
			}
		}
		fmt.Fprintf(&b, "    %s {\n", mermaidID(t.ModelName))
		any := false
		for _, c := range t.Columns {
			if !pk[c.Name] && !fk[c.Name] {
				continue
			}
			tag := ""
			switch {
			case pk[c.Name] && fk[c.Name]:
				tag = "PK, FK"
			case pk[c.Name]:
				tag = "PK"
			default:
				tag = "FK"
			}
			fmt.Fprintf(&b, "        %s %s %s\n", erType(c), safeIdent(c.Name), tag)
			any = true
		}
		if !any {
			b.WriteString("        string id PK\n")
		}
		b.WriteString("    }\n")
	}
	for _, name := range sortedKeys(stubs) {
		fmt.Fprintf(&b, "    %s {\n        string externalStub PK\n    }\n", mermaidID(name))
	}
	for _, t := range sorted {
		for _, f := range t.ForeignKeys {
			fmt.Fprintf(&b, "    %s }o--|| %s : \"%s\"\n",
				mermaidID(t.ModelName), mermaidID(f.ReferencedModel), f.Column)
		}
	}
	b.WriteString("```\n\n")
	return b.String()
}

// writeModelDetail emits per-model field tables and the enum list for a leaf.
func writeModelDetail(b *strings.Builder, tables []*schema.Table, enums []*schema.Enum) {
	for _, t := range tables {
		fmt.Fprintf(b, "### `%s`", t.ModelName)
		if t.Name != "" {
			fmt.Fprintf(b, " → `%s`", t.Name)
		}
		b.WriteString("\n\n")
		if t.Comment != "" {
			fmt.Fprintf(b, "%s\n\n", escapeCell(t.Comment))
		}
		b.WriteString("| Column | Type | Null |\n| --- | --- | --- |\n")
		for _, c := range t.Columns {
			null := "not null"
			if c.Optional {
				null = "nullable"
			}
			fmt.Fprintf(b, "| `%s` | `%s` | %s |\n", c.Name, columnType(c), null)
		}
		b.WriteString("\n")
	}
	if len(enums) > 0 {
		b.WriteString("### Enums\n\n")
		for _, e := range enums {
			vals := make([]string, len(e.Values))
			for i, v := range e.Values {
				vals[i] = v.Name
			}
			fmt.Fprintf(b, "- `%s`: %s\n", e.Name, strings.Join(vals, ", "))
		}
		b.WriteString("\n")
	}
}

func columnType(c *schema.Column) string {
	if c.Enum != nil {
		return c.Enum.Name
	}
	return c.SQLType
}

// erType maps a column to a Mermaid-safe attribute type keyword.
func erType(c *schema.Column) string {
	if c.Enum != nil {
		return "string"
	}
	base, _ := types.BaseType(c.SQLType)
	switch base {
	case "BOOLEAN", "BOOL":
		return "bool"
	case "SMALLINT", "INTEGER", "INT", "BIGINT":
		return "int"
	case "REAL", "DOUBLE PRECISION", "NUMERIC", "DECIMAL":
		return "float"
	case "TIMESTAMPTZ", "TIMESTAMP", "DATE", "TIME":
		return "datetime"
	case "JSON", "JSONB":
		return "json"
	case "BYTEA":
		return "bytes"
	default:
		return "string"
	}
}

func mermaidID(name string) string {
	id := safeIdent(name)
	if id == "" {
		return "Model"
	}
	return id
}

func safeIdent(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	return b.String()
}

func escapeCell(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "|", "\\|"), "\n", " ")
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
