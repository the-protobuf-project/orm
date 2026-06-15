package generator

// config.go loads the optional protorm.yaml layout config (passed via the
// `config=<path>` plugin option) and resolves a proto package to its target
// database and postgres schema. This is what lets a multi-service monorepo split
// into several databases with clean schema names without annotating every file.
// Precedence: a per-file (protorm.v1.datasource) annotation wins over the
// config, which in turn wins over the package-path defaults.

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// layoutConfig is the parsed protorm.yaml.
type layoutConfig struct {
	Datasources []matchRule `yaml:"datasources"`
}

// matchRule assigns every proto package matching Match to a database and schema.
// Match is a dotted glob over the package: a trailing "**" matches any remaining
// segments ("fleet.tracking.**"); "*" matches exactly one segment; other
// segments match literally.
type matchRule struct {
	Match    string `yaml:"match"`
	Database string `yaml:"database"`
	// Schema is a literal or a template using {leaf} (the last package segment
	// with a trailing API version dropped). Empty falls back to SchemaDepth.
	Schema string `yaml:"schema"`
	// SchemaDepth, when >0 and Schema is empty, joins the first N package
	// segments with "_" to form the schema name.
	SchemaDepth int `yaml:"schema_depth"`
}

// loadLayoutConfig reads protorm.yaml from path. A blank path yields nil (no
// config; defaults apply).
func loadLayoutConfig(path string) (*layoutConfig, error) {
	if path == "" {
		return nil, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}
	var c layoutConfig
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("parse config %q: %w", path, err)
	}
	return &c, nil
}

// resolve returns the database and schema for a proto package under the first
// matching rule, or empty strings when nothing matches (nil-safe).
func (c *layoutConfig) resolve(pkg string) (database, schema string) {
	if c == nil {
		return "", ""
	}
	segs := strings.Split(pkg, ".")
	for _, r := range c.Datasources {
		if matchPackage(r.Match, segs) {
			return r.Database, ruleSchema(r, segs)
		}
	}
	return "", ""
}

// ruleSchema computes the schema name a matched rule assigns to a package.
func ruleSchema(r matchRule, segs []string) string {
	switch {
	case r.Schema != "":
		return strings.ReplaceAll(r.Schema, "{leaf}", leafSegment(segs))
	case r.SchemaDepth > 0 && r.SchemaDepth <= len(segs):
		return strings.Join(segs[:r.SchemaDepth], "_")
	default:
		return ""
	}
}

// leafSegment returns the last package segment, dropping a trailing API version
// ("store.apps.calendar.v1" → "calendar").
func leafSegment(segs []string) string {
	i := len(segs) - 1
	if i > 0 && isVersionSegment(segs[i]) {
		i--
	}
	if i < 0 {
		return ""
	}
	return segs[i]
}

// matchPackage reports whether the dotted glob pattern matches package segments.
func matchPackage(pattern string, segs []string) bool {
	pats := strings.Split(pattern, ".")
	for i, p := range pats {
		if p == "**" {
			return true // trailing wildcard: the rest matches
		}
		if i >= len(segs) {
			return false
		}
		if p != "*" && p != segs[i] {
			return false
		}
	}
	return len(pats) == len(segs)
}
