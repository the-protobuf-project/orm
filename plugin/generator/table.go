package generator

// table.go assembles a *schema.Table from a proto message: its column list,
// synthesized id/timestamp columns, and the foreign keys inferred from
// resource_reference. Scalar/enum column mapping lives in column.go.

import (
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/oh-tarnished/protorm/plugin/generator/schema"
	"github.com/oh-tarnished/protorm/protorm/protormpbv1"
)

// buildTable maps one resource-annotated message to a *schema.Table.
func (ctx *buildCtx) buildTable(db *schema.Database, s *schema.Schema, msg *protogen.Message, name, src, srcPath string) *schema.Table {
	t := &schema.Table{
		Name:         name,
		Comment:      cleanComment(msg.Comments.Leading),
		ModelName:    string(msg.Desc.Name()),
		ProtoMessage: string(msg.Desc.FullName()),
		SourceFile:   src,
		SourceProto:  srcPath,
	}

	ctx.populateColumns(db, s, t, msg)
	applyAIPSystemFields(t)
	if res := resourceOf(msg); res != nil {
		materializeParents(t, res)
	}

	tOpts := tableOpts(msg)
	applyIDStrategy(t, idStrategyOf(tOpts))
	applyTimestamps(t, tOpts.GetTimestamps())

	for _, idx := range tOpts.GetIndexes() {
		t.Indexes = append(t.Indexes, &schema.Index{
			Name: idx.GetIndex(), Columns: idx.GetColumns(), Unique: idx.GetUnique(),
		})
	}
	return t
}

// populateColumns maps msg's fields onto t. Scalar/enum fields become columns;
// string fields with google.api.resource_reference become FK columns; and
// user message-typed fields become embed requests (normalized into related
// tables by normalizeEmbeds) instead of lossy JSONB blobs — unless the field
// is skipped or pins an explicit column type. Shared by buildTable (resources)
// and materialize (embedded children).
func (ctx *buildCtx) populateColumns(db *schema.Database, s *schema.Schema, t *schema.Table, msg *protogen.Message) {
	for _, f := range msg.Fields {
		cOpts := colOpts(f)
		if target := normalizableMessage(f); target != "" && cOpts.GetType() == "" {
			if cOpts.GetSkip() {
				continue
			}
			// storage=json inlines a message field as a single JSONB column (a
			// value object or metadata blob) instead of relationalizing it into a
			// child table. Default (unset/relation) keeps the lossless embed.
			if cOpts.GetStorage() == protormpbv1.StorageMode_STORAGE_MODE_JSON {
				notNull := isRequiredField(f)
				t.Columns = append(t.Columns, &schema.Column{
					Name:     colName(f, cOpts),
					Comment:  cleanComment(f.Comments.Leading),
					SQLType:  "JSONB",
					NotNull:  notNull,
					Optional: !notNull,
				})
				continue
			}
			ctx.embeds = append(ctx.embeds, &embedReq{
				db: db, schemaName: s.Name, parent: t, field: f,
				targetMsg: target, repeated: f.Desc.IsList(),
				optional: !isRequiredField(f),
				onDelete: refAction(cOpts.GetOnDelete()),
				onUpdate: refAction(cOpts.GetOnUpdate()),
			})
			continue
		}

		col := buildColumn(s, f)
		if col == nil {
			continue
		}
		t.Columns = append(t.Columns, col)
		if col.PrimaryKey && t.PKColumn == "" {
			t.PKColumn = col.Name
		}
		if ref := resourceRef(f); ref != nil {
			refSchema, refTable := schemaTable(ref.GetType(), "")
			refModel := modelNameFromType(ref.GetType())
			col.FKModel = refModel
			t.ForeignKeys = append(t.ForeignKeys, &schema.ForeignKey{
				Column:           col.Name,
				ReferencedSchema: refSchema,
				ReferencedTable:  refTable,
				ReferencedModel:  refModel,
				OnDelete:         refAction(cOpts.GetOnDelete()),
				OnUpdate:         refAction(cOpts.GetOnUpdate()),
				// ReferencedColumn filled by resolveRelations after all tables built.
			})
		}
	}
	ctx.addOneofDiscriminators(s, t, msg)
}

// refAction converts a ReferentialAction enum to its SQL clause form.
func refAction(a protormpbv1.ReferentialAction) string {
	switch a {
	case protormpbv1.ReferentialAction_REFERENTIAL_ACTION_CASCADE:
		return "CASCADE"
	case protormpbv1.ReferentialAction_REFERENTIAL_ACTION_RESTRICT:
		return "RESTRICT"
	case protormpbv1.ReferentialAction_REFERENTIAL_ACTION_SET_NULL:
		return "SET NULL"
	case protormpbv1.ReferentialAction_REFERENTIAL_ACTION_SET_DEFAULT:
		return "SET DEFAULT"
	case protormpbv1.ReferentialAction_REFERENTIAL_ACTION_NO_ACTION:
		return "NO ACTION"
	default:
		return ""
	}
}
