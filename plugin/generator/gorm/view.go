package gorm

// view.go prepares the models.go template view: naming, Go types, struct tags,
// enum consts, and conditional imports. The template is presentation only.

import (
	"strings"

	"github.com/oh-tarnished/protorm/plugin/generator/header"
	"github.com/oh-tarnished/protorm/plugin/generator/naming"
	"github.com/oh-tarnished/protorm/plugin/generator/schema"
)

type fieldView struct{ Comment, Decl string }

type modelView struct {
	Comment, Name, TableName string
	Fields                   []fieldView
}

type enumValueView struct{ ConstName, TypeName, MapName string }

type enumView struct {
	Comment, Name string
	Values        []enumValueView
}

// packageView assembles the template data for one schema package.
func packageView(db *schema.Database, s *schema.Schema, pkg string) map[string]any {
	var models []modelView
	needTime, needJSON := false, false

	for _, t := range s.Tables {
		m := modelView{
			Comment:   commentOr(t.Comment, t.ModelName+" model."),
			Name:      t.ModelName,
			TableName: s.Name + "." + t.Name,
		}
		// Reserve scalar Go field names so association fields stay unique — two
		// FKs to the same model must not produce two identically-named fields.
		used := map[string]bool{}
		for _, col := range t.Columns {
			used[naming.PascalGo(col.Name)] = true
		}
		for _, col := range t.Columns {
			gt := goType(col)
			needTime = needTime || strings.Contains(gt, "time.Time")
			needJSON = needJSON || strings.Contains(gt, "json.RawMessage")

			goField := naming.PascalGo(col.Name)
			m.Fields = append(m.Fields, fieldView{
				Comment: col.Comment,
				Decl:    goField + " " + gt + " `" + structTag(col) + "`",
			})
			// BelongsTo association: emitted alongside the FK column. The field is
			// named after the FK column (minus _id) so multiple references to the
			// same model stay distinct; GORM resolves the link via foreignKey.
			if col.FKModel != "" {
				assoc := uniqueGoName(naming.PascalGo(naming.StripIDSuffix(col.Name)), used)
				m.Fields = append(m.Fields, fieldView{
					Decl: assoc + " *" + col.FKModel +
						" `gorm:\"foreignKey:" + goField + constraintTag(t, col.Name) +
						"\" json:\"" + strings.ToLower(assoc) + ",omitempty\"`",
				})
			}
		}
		// HasMany back-references (e.g. Author.Books []Book).
		for _, hm := range t.HasMany {
			field := uniqueGoName(naming.PascalGo(hm.Field), used)
			m.Fields = append(m.Fields, fieldView{
				Comment: "Back-relation: " + hm.Model + " records that reference this via " + hm.ViaFK + ".",
				Decl: field + " []" + hm.Model +
					" `gorm:\"foreignKey:" + naming.PascalGo(hm.ViaFK) + "\" json:\"" + strings.ToLower(field) + ",omitempty\"`",
			})
		}
		models = append(models, m)
	}

	var imports []string
	if needJSON {
		imports = append(imports, "encoding/json")
	}
	if needTime {
		imports = append(imports, "time")
	}

	return map[string]any{
		"Header": header.Render("//", header.Info{
			PluginVersion: db.PluginVersion,
			ProtocVersion: db.ProtocVersion,
			Source:        strings.Join(s.SourceProtos(), ", "),
			Database:      db.Name,
			Schema:        s.Name,
		}),
		"Package": pkg,
		"Imports": imports,
		"Enums":   enumViews(s),
		"Models":  models,
	}
}

// enumViews renders each schema enum as a Go string type with one const per value.
func enumViews(s *schema.Schema) []enumView {
	var out []enumView
	for _, e := range s.Enums {
		ev := enumView{
			Comment: commentOr(e.Comment, e.Name+" enumerates the "+e.SQLName+" values."),
			Name:    e.Name,
		}
		for _, v := range e.Values {
			ev.Values = append(ev.Values, enumValueView{
				ConstName: e.Name + naming.PascalGo(strings.ToLower(v.Name)),
				TypeName:  e.Name,
				MapName:   v.MapName,
			})
		}
		out = append(out, ev)
	}
	return out
}
