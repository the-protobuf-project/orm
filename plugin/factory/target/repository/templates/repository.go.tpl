{{.Header}}

package {{.Package}}

import (
{{- range .Imports}}
	{{.}}
{{- end}}
)

{{- range .Resources}}

// {{.Model}}Repository is the provider-agnostic persistence surface for
// {{.Model}} ({{.Pattern}}): full-message writes, AIP-160/AIP-132 lists with
// opaque page tokens, field-mask updates, and etag optimistic concurrency.
type {{.Model}}Repository interface {
{{- if .Parented}}
	// Create persists m under parent and returns the stored record.
	Create(ctx context.Context, parent string, m *{{.PB}}) (*{{.PB}}, error)
{{- else}}
	// Create persists m and returns the stored record.
	Create(ctx context.Context, m *{{.PB}}) (*{{.PB}}, error)
{{- end}}
	// Get returns the record addressed by its resource name, or
	// repox.ErrNotFound.
	Get(ctx context.Context, name string) (*{{.PB}}, error)
{{- if .Parented}}
	// List returns one page of records under parent.
	List(ctx context.Context, parent string, in repox.ListInput) ([]*{{.PB}}, string, error)
{{- else}}
	// List returns one page of records.
	List(ctx context.Context, in repox.ListInput) ([]*{{.PB}}, string, error)
{{- end}}
	// Update persists the masked fields of m; an empty mask replaces every
	// mutable field. m.Etag, when set, guards against concurrent writes
	// (repox.ErrConflict).
	Update(ctx context.Context, m *{{.PB}}, paths []string) (*{{.PB}}, error)
	// Delete removes the record addressed by its resource name.
	Delete(ctx context.Context, name string) error
}

// {{.Model}}Hooks lets custom logic run inside the generated {{.Model}}
// adapters without editing them: normalization/derivation before writes,
// derived fields after reads, and guards before deletes. Nil funcs are skipped.
type {{.Model}}Hooks struct {
	// BeforeCreate runs after name/id resolution, before the row is written.
	BeforeCreate func(ctx context.Context, m *{{.PB}}) error
	// AfterRead runs on every record a read returns (Get, List, and the
	// re-reads writes return).
	AfterRead func(ctx context.Context, m *{{.PB}}) error
	// BeforeUpdate runs after the mask is applied to the merged record,
	// before it is written.
	BeforeUpdate func(ctx context.Context, existing, merged *{{.PB}}, paths []string) error
	// BeforeDelete runs before the row is removed; returning an error vetoes.
	BeforeDelete func(ctx context.Context, name string) error
}
{{- end}}
