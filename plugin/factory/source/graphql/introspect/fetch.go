package introspect

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Fetch runs the introspection query against endpoint and returns the decoded
// Schema. Each entry in headers is sent on the request (e.g. an admin secret or
// bearer token). The endpoint must be the full GraphQL URL (e.g.
// "http://localhost:3280/graphql").
func Fetch(ctx context.Context, endpoint string, headers map[string]string) (*Schema, error) {
	body, err := json.Marshal(map[string]string{"query": introspectionQuery})
	if err != nil {
		return nil, fmt.Errorf("failed to encode introspection query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to build introspection request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("introspection request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read introspection response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("introspection returned status %d: %s", resp.StatusCode, truncate(raw, 256))
	}

	return Decode(raw)
}

// Decode parses introspection JSON into a Schema. It accepts both the raw server
// response envelope ({"data":{"__schema":...}}) and a bare Schema object (as written
// by the introspect command's -o output), so cached files round-trip cleanly.
func Decode(raw []byte) (*Schema, error) {
	var out Response
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("failed to decode introspection JSON: %w", err)
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("introspection returned errors: %s", out.Errors[0].Message)
	}
	if out.Data.Schema.QueryType != nil {
		return &out.Data.Schema, nil
	}

	// Fall back to a bare Schema object.
	var bare Schema
	if err := json.Unmarshal(raw, &bare); err == nil && bare.QueryType != nil {
		return &bare, nil
	}
	return nil, fmt.Errorf("introspection response missing __schema (is this a GraphQL endpoint?)")
}

// truncate returns at most n bytes of b as a string, for safe error messages.
func truncate(b []byte, n int) string {
	if len(b) > n {
		return string(b[:n])
	}
	return string(b)
}
