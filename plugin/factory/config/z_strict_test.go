package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/the-protobuf-project/orm/plugin/factory/config"
)

func TestStrict_InlineBackendKeys(t *testing.T) {
	y := `
version: 1
strip_version: true
dedupe_schema_table: true
datasources:
  - {match: ".*", database: mydb, schema: public, schema_depth: 1, strip_version: true}
otel:
  enabled: true
  metrics: true
graphql:
  endpoint: http://x/graphql
`
	p := filepath.Join(t.TempDir(), "orm.yaml")
	if err := os.WriteFile(p, []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(p)
	if err != nil {
		t.Fatalf("existing orm.yaml with backend keys failed strict decode: %v", err)
	}
	if cfg.GraphQL == nil || cfg.GraphQL.Endpoint == "" {
		t.Fatalf("graphql not parsed")
	}
	t.Logf("OK datasources=%d otel=%v", len(cfg.Datasources), cfg.OTel != nil)
}
