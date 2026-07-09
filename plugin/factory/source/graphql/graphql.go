// Package graphql is the factory Source that reads a GraphQL server: it
// introspects a live endpoint (or decodes a cached schema.json), then builds the
// GraphQL IR under the configured dialect. Unlike the proto Source it runs
// outside protoc — its input is an endpoint or a file, supplied via Config — so
// it drives the CLI (`protoc-gen-orm graphql …`) rather than plugin mode.
package graphql

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/the-protobuf-project/orm/plugin/factory"
	"github.com/the-protobuf-project/orm/plugin/factory/coreir"
	"github.com/the-protobuf-project/orm/plugin/factory/source/graphql/dialect"
	"github.com/the-protobuf-project/orm/plugin/factory/source/graphql/introspect"
	"github.com/the-protobuf-project/orm/plugin/factory/source/graphql/ir"
)

// Config selects the introspection input and conventions for a GraphQL build.
type Config struct {
	// Endpoint is a live GraphQL URL to introspect (used when SchemaFile is empty).
	Endpoint string
	// SchemaFile is a cached introspection JSON file (skips the live fetch).
	SchemaFile string
	// AdminSecret, when set, is sent under the dialect's auth header.
	AdminSecret string
	// Headers are extra "Key: Value" request headers.
	Headers []string
	// Dialect supplies engine conventions; defaults to dialect.Default().
	Dialect dialect.Dialect
}

// Source builds the GraphQL IR from an endpoint or cached schema.
type Source struct{ cfg Config }

// New returns a GraphQL source. A nil Dialect defaults to dialect.Default().
func New(cfg Config) *Source {
	if cfg.Dialect == nil {
		cfg.Dialect = dialect.Default()
	}
	return &Source{cfg: cfg}
}

// Name identifies this source in the registry and config.
func (s *Source) Name() string { return "graphql" }

// Build loads the introspection schema and normalizes it to the IR.
func (s *Source) Build(_ factory.Ctx) (*coreir.Model, error) {
	schema, err := s.loadSchema()
	if err != nil {
		return nil, err
	}
	return &coreir.Model{
		GraphQL:    ir.Build(schema, s.cfg.Dialect),
		RawGraphQL: schema,
	}, nil
}

// loadSchema reads the introspection schema from SchemaFile or fetches it live.
func (s *Source) loadSchema() (*introspect.Schema, error) {
	if s.cfg.SchemaFile != "" {
		raw, err := os.ReadFile(s.cfg.SchemaFile)
		if err != nil {
			return nil, fmt.Errorf("read schema file: %w", err)
		}
		return introspect.Decode(raw)
	}
	if s.cfg.Endpoint == "" {
		return nil, fmt.Errorf("graphql source needs an endpoint or a schema file")
	}
	headers := map[string]string{}
	if s.cfg.AdminSecret != "" {
		if h := s.cfg.Dialect.AuthHeader(); h != "" {
			headers[h] = s.cfg.AdminSecret
		}
	}
	for _, h := range s.cfg.Headers {
		if key, value, ok := strings.Cut(h, ":"); ok {
			headers[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
	}
	return introspect.Fetch(context.Background(), s.cfg.Endpoint, headers)
}
