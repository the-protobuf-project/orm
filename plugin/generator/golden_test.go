package generator_test

// Golden-file tests for the database backends. Every directory under
// testdata/cases/ is one case: its .proto files are compiled in-process (via the
// protokit golden harness) and each database target's output is compared
// byte-for-byte against <case>/golden/<target>/. RunCase skips any target
// this module's registry doesn't ship.
//
// Regenerate goldens after an intentional output change:
//
//	go test ./plugin/generator -run TestGolden -update

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/the-protobuf-project/orm/plugin/generator"
	"github.com/the-protobuf-project/orm/plugin/generator/backend"
	"github.com/the-protobuf-project/protokit/golden"
	"github.com/the-protobuf-project/protokit/header"
	"github.com/the-protobuf-project/protokit/schema"
)

// defaultTargets are the database backends every golden case runs unless it
// ships a "targets" file.
var defaultTargets = []string{"gorm", "prisma", "sql"}

// ormBackend builds the orm Backend for one golden case: it reads any orm.yaml
// the case ships (grouping/otel config) and its optional "stores"/"converters"
// markers, and mirrors the binary's opt defaults (go_module set so the gorm
// aggregator is emitted, otel on). protokit's harness stays generator-neutral —
// all of this generator-specific knowledge lives here, not in RunCase.
func ormBackend(dir string) schema.Backend {
	var cfg *backend.Config
	if path := filepath.Join(dir, "orm.yaml"); fileExists(path) {
		c, err := backend.LoadConfig(path)
		if err != nil {
			panic(err) // a malformed fixture config is a test-authoring bug
		}
		cfg = c
	}
	stores := fileExists(filepath.Join(dir, "stores"))
	converters := fileExists(filepath.Join(dir, "converters"))
	return backend.New(cfg, "example.com/test/gen", stores, true, converters)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// TestMain stamps the orm tool name into generated banners so goldens match what
// the protoc-gen-orm binary produces (protokit's generator-neutral default names
// the framework instead).
func TestMain(m *testing.M) {
	header.SetTool("protoc-gen-orm")
	os.Exit(m.Run())
}

func TestGolden(t *testing.T) {
	cases, err := os.ReadDir("testdata/cases")
	if err != nil {
		t.Fatalf("read cases: %v", err)
	}
	for _, c := range cases {
		if !c.IsDir() {
			continue
		}
		t.Run(c.Name(), func(t *testing.T) {
			golden.RunCase(t, filepath.Join("testdata", "cases", c.Name()), generator.Targets(), defaultTargets, ormBackend)
		})
	}
}
