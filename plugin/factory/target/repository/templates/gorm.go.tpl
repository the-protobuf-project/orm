{{.Header}}

package {{.Package}}

import (
{{- range .Imports}}
	{{.}}
{{- end}}
)

{{- range .Resources}}

// Gorm{{.Model}}Repository is the GORM adapter of {{.Model}}Repository: flat
// CRUD over the generated model, store, converters, and filterx spec. Exported
// fields and helpers are the Tier-2 seam — embed it and override whole methods
// for behavior the schema cannot express; generated methods only call private
// helpers, never the public interface, so overrides never re-enter generated
// code.
type Gorm{{.Model}}Repository struct {
	DB    *gorm.DB
	Hooks {{.Model}}Hooks
	// ListOverrides substitute the generated dispatch for single filter fields
	// (e.g. a derived state computed via subqueries).
	ListOverrides map[string]filterx.SQLHandler
}

// NewGorm{{.Model}}Repository returns the GORM adapter bound to db.
func NewGorm{{.Model}}Repository(db *gorm.DB) *Gorm{{.Model}}Repository {
	return &Gorm{{.Model}}Repository{DB: db}
}

{{if .Parented}}
// Create persists in under parent and returns the stored record.
func (r *Gorm{{.Model}}Repository) Create(ctx context.Context, parent string, in *{{.PB}}) (*{{.PB}}, error) {
	parentIDs, err := repox.SplitName(parent, {{.ParentCollExpr}})
	if err != nil {
		return nil, err
	}
	id := repox.NewULID()
	if in.GetName() != "" {
		ids, err := repox.SplitName(in.GetName(), {{.CollectionsExpr}})
		if err != nil {
			return nil, err
		}
		for i := range parentIDs {
			if ids[i] != parentIDs[i] {
				return nil, fmt.Errorf("%w: name %q is not under parent %q", repox.ErrInvalidArgument, in.GetName(), parent)
			}
		}
		id = ids[len(ids)-1]
	}
	in = proto.Clone(in).(*{{.PB}})
	in.Name = {{.FormatCallParented}}
	if h := r.Hooks.BeforeCreate; h != nil {
		if err := h(ctx, in); err != nil {
			return nil, err
		}
	}
	m := gormdb.{{.Model}}FromProto(in)
	m.{{.PKField}} = id
	m.Name = in.GetName()
	m.{{.ParentFKField}} = parentIDs[len(parentIDs)-1]
	{{- range .RefsCreate}}
	{{.}}
	{{- end}}
	{{- if .HasEtag}}
	m.{{.EtagField}} = {{if .EtagPtr}}repox.Ptr(repox.NewULID()){{else}}repox.NewULID(){{end}}
	{{- end}}
	if err := gormdb.{{.Store}}(r.DB).Create(ctx, m); err != nil {
		return nil, repox.MapGormErr(err)
	}
	return r.get(ctx, id)
}
{{else}}
// Create persists in and returns the stored record.
func (r *Gorm{{.Model}}Repository) Create(ctx context.Context, in *{{.PB}}) (*{{.PB}}, error) {
	id := repox.NewULID()
	if in.GetName() != "" {
		ids, err := repox.SplitName(in.GetName(), {{.CollectionsExpr}})
		if err != nil {
			return nil, err
		}
		id = ids[len(ids)-1]
	}
	in = proto.Clone(in).(*{{.PB}})
	in.Name = Format{{.Model}}Name(id)
	if h := r.Hooks.BeforeCreate; h != nil {
		if err := h(ctx, in); err != nil {
			return nil, err
		}
	}
	m := gormdb.{{.Model}}FromProto(in)
	m.{{.PKField}} = id
	m.Name = in.GetName()
	{{- range .RefsCreate}}
	{{.}}
	{{- end}}
	{{- if .HasEtag}}
	m.{{.EtagField}} = {{if .EtagPtr}}repox.Ptr(repox.NewULID()){{else}}repox.NewULID(){{end}}
	{{- end}}
	if err := gormdb.{{.Store}}(r.DB).Create(ctx, m); err != nil {
		return nil, repox.MapGormErr(err)
	}
	return r.get(ctx, id)
}
{{end}}

// Get returns the record addressed by its resource name.
func (r *Gorm{{.Model}}Repository) Get(ctx context.Context, name string) (*{{.PB}}, error) {
	ids, err := repox.SplitName(name, {{.CollectionsExpr}})
	if err != nil {
		return nil, err
	}
	return r.get(ctx, ids[len(ids)-1])
}

// get loads by surrogate key — the private read every generated method re-reads
// through, so Tier-2 overrides of Get never re-enter generated writes.
func (r *Gorm{{.Model}}Repository) get(ctx context.Context, id string) (*{{.PB}}, error) {
	var m gormdb.{{.Model}}
	if err := r.DB.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, repox.MapGormErr(err)
	}
	return r.toProto(ctx, &m)
}

// toProto converts a loaded row, decorating reference names and running the
// AfterRead hook.
func (r *Gorm{{.Model}}Repository) toProto(ctx context.Context, m *gormdb.{{.Model}}) (*{{.PB}}, error) {
	out := gormdb.{{.Model}}ToProto(m)
	{{- range .RefsToProto}}
	{{.}}
	{{- end}}
	if h := r.Hooks.AfterRead; h != nil {
		if err := h(ctx, out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

{{if .Parented}}
// List returns one page of records under parent.
func (r *Gorm{{.Model}}Repository) List(ctx context.Context, parent string, in repox.ListInput) ([]*{{.PB}}, string, error) {
	parentIDs, err := repox.SplitName(parent, {{.ParentCollExpr}})
	if err != nil {
		return nil, "", err
	}
	scope := r.DB.Where("{{.ParentFKColumn}} = ?", parentIDs[len(parentIDs)-1])
	return r.list(ctx, scope, in)
}
{{else}}
// List returns one page of records.
func (r *Gorm{{.Model}}Repository) List(ctx context.Context, in repox.ListInput) ([]*{{.PB}}, string, error) {
	return r.list(ctx, r.DB, in)
}
{{end}}

func (r *Gorm{{.Model}}Repository) list(ctx context.Context, scope *gorm.DB, in repox.ListInput) ([]*{{.PB}}, string, error) {
	conds, err := filterx.Parse(in.Filter)
	if err != nil {
		return nil, "", repox.MapFilterxErr(err)
	}
	eng := filterx.Gorm[gormdb.{{.Model}}](gormdb.{{.Model}}FilterSpec)
	for f, h := range r.ListOverrides {
		eng.Override(f, h)
	}
	rows, next, err := eng.List(ctx, scope, filterx.ListInput{
		PageSize:  in.PageSize,
		PageToken: in.PageToken,
		OrderBy:   in.OrderBy,
		Filter:    conds,
	})
	if err != nil {
		return nil, "", repox.MapFilterxErr(err)
	}
	items := make([]*{{.PB}}, 0, len(rows))
	for i := range rows {
		out, err := r.toProto(ctx, &rows[i])
		if err != nil {
			return nil, "", err
		}
		items = append(items, out)
	}
	return items, next, nil
}

// Update persists the masked fields of in; an empty mask replaces every
// mutable field. The write happens in one transaction guarded by in.Etag.
func (r *Gorm{{.Model}}Repository) Update(ctx context.Context, in *{{.PB}}, paths []string) (*{{.PB}}, error) {
	ids, err := repox.SplitName(in.GetName(), {{.CollectionsExpr}})
	if err != nil {
		return nil, err
	}
	id := ids[len(ids)-1]
	err = r.DB.Transaction(func(tx *gorm.DB) error {
		var existing gormdb.{{.Model}}
		if err := tx.WithContext(ctx).First(&existing, "id = ?", id).Error; err != nil {
			return err
		}
		{{- if .HasEtag}}
		if in.GetEtag() != "" && {{if .EtagPtr}}existing.{{.EtagField}} != nil && *existing.{{.EtagField}}{{else}}existing.{{.EtagField}}{{end}} != in.GetEtag() {
			return repox.ErrConflict
		}
		{{- end}}
		existingPB := gormdb.{{.Model}}ToProto(&existing)
		{{- if .RefsToProto}}
		{
			m, out := &existing, existingPB
			{{- range .RefsToProto}}
			{{.}}
			{{- end}}
		}
		{{- end}}
		merged := proto.Clone(existingPB).(*{{.PB}})
		apply{{.Model}}Mask(merged, in, paths)
		if h := r.Hooks.BeforeUpdate; h != nil {
			if err := h(ctx, existingPB, merged, paths); err != nil {
				return err
			}
		}
		next := gormdb.{{.Model}}FromProto(merged)
		_ = next
		{{- range .MutableAssigns}}
		{{.}}
		{{- end}}
		{{- if .HasEtag}}
		existing.{{.EtagField}} = {{if .EtagPtr}}repox.Ptr(repox.NewULID()){{else}}repox.NewULID(){{end}}
		{{- end}}
		return gormdb.{{.Store}}(tx).Update(ctx, &existing)
	})
	if err != nil {
		return nil, repox.MapGormErr(err)
	}
	return r.get(ctx, id)
}

// Delete removes the record addressed by its resource name.
func (r *Gorm{{.Model}}Repository) Delete(ctx context.Context, name string) error {
	ids, err := repox.SplitName(name, {{.CollectionsExpr}})
	if err != nil {
		return err
	}
	if h := r.Hooks.BeforeDelete; h != nil {
		if err := h(ctx, name); err != nil {
			return err
		}
	}
	id := ids[len(ids)-1]
	var existing gormdb.{{.Model}}
	if err := r.DB.WithContext(ctx).First(&existing, "id = ?", id).Error; err != nil {
		return repox.MapGormErr(err)
	}
	return repox.MapGormErr(gormdb.{{.Store}}(r.DB).DeleteByID(ctx, id))
}

// Compile-time proof the adapter satisfies the interface.
var _ {{.Model}}Repository = (*Gorm{{.Model}}Repository)(nil)
{{- end}}
