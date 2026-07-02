package gorm

import (
	"embed"

	"github.com/the-protobuf-project/protokit/templates"
)

//go:embed templates/*.tpl
var templateFS embed.FS

// tmpl is this target's own parsed template set, keyed by file base name.
var tmpl = templates.MustParse(templateFS, "templates/*.tpl")
