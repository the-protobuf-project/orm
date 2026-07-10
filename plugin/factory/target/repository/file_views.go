package repository

// file_views.go assembles the per-file template data for names.go, mask.go,
// and gorm.go from the planned gormResourceViews. Each view carries its own
// import list so every emitted file imports exactly what its fragments use.

import (
	"strings"

	"github.com/the-protobuf-project/protokit/header"
	"github.com/the-protobuf-project/protokit/schema"
)

// fileHeader renders the banner for one generated file of schema s.
func fileHeader(db *schema.Database, s *schema.Schema, note string) string {
	return header.Render("//", header.Info{
		PluginVersion: db.PluginVersion,
		ProtocVersion: db.ProtocVersion,
		Source:        strings.Join(s.SourceProtos(), ", "),
		Database:      db.Name,
		Schema:        s.Name,
		Notes:         []string{note},
	})
}

// namesView prepares names.go: the per-resource resource-name codecs.
func namesView(db *schema.Database, s *schema.Schema, pkg string, rs []gormResourceView) map[string]any {
	return map[string]any{
		"Header":      fileHeader(db, s, "AIP resource-name codecs for this schema's repositories."),
		"Package":     pkg,
		"RepoxImport": dbGoModule(db) + "/" + repoxPkg,
		"Resources":   rs,
	}
}

// maskView prepares mask.go: the per-resource field-mask merge functions.
func maskView(pb *pbIndex, db *schema.Database, s *schema.Schema, pkg string, rs []gormResourceView) map[string]any {
	imports := map[string]string{
		dbGoModule(db) + "/" + repoxPkg: "",
	}
	addPBImports(pb, s, imports)
	return map[string]any{
		"Header":    fileHeader(db, s, "Field-mask merge functions shared by every adapter of this schema."),
		"Package":   pkg,
		"Imports":   renderImports(imports),
		"Resources": rs,
	}
}

// gormFileView prepares gorm.go: the GORM adapters.
func gormFileView(pb *pbIndex, db *schema.Database, s *schema.Schema, pkg string, rs []gormResourceView) map[string]any {
	imports := map[string]string{
		"context":                                    "",
		"gorm.io/gorm":                               "",
		"google.golang.org/protobuf/proto":           "",
		dbGoModule(db) + "/" + repoxPkg:              "",
		dbGormModule(db) + "/filterx":                "",
		dbGormModule(db) + "/" + db.Name + "/" + pkg: "gormdb",
	}
	for _, r := range rs {
		if r.Parented {
			imports["fmt"] = ""
			break
		}
	}
	addPBImports(pb, s, imports)
	return map[string]any{
		"Header":    fileHeader(db, s, "GORM adapters composing the generated models, stores, converters, and filterx specs."),
		"Package":   pkg,
		"Imports":   renderImports(imports),
		"Resources": rs,
	}
}

// addPBImports adds the generated proto packages of s's repository resources.
func addPBImports(pb *pbIndex, s *schema.Schema, imports map[string]string) {
	for _, t := range s.Tables {
		if t.Source == nil {
			continue
		}
		if msg, ok := pb.msgs[t.Source.FullName()]; ok && resourcePattern(t.Source) != "" && !t.ValueObject {
			path := string(msg.GoIdent.GoImportPath)
			imports[path] = goPackageName(path)
		}
	}
}
