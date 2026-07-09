package ir

import "github.com/the-protobuf-project/orm/plugin/factory/source/graphql/introspect"

// resolve flattens a (possibly wrapped) introspection TypeRef into a FieldType,
// recording whether the outer type is non-null, whether it is a list, and whether
// list elements are non-null. It handles the common single-level-list case used by
// GraphQL schemas; deeper nesting collapses to the innermost named type.
func resolve(ref introspect.TypeRef) FieldType {
	ft := FieldType{}
	cur := &ref

	if cur.Kind == "NON_NULL" {
		ft.NonNull = true
		cur = cur.OfType
	}
	if cur != nil && cur.Kind == "LIST" {
		ft.List = true
		cur = cur.OfType
		if cur != nil && cur.Kind == "NON_NULL" {
			ft.ElemNonNull = true
			cur = cur.OfType
		}
	}
	// Descend any remaining wrappers to reach the named type.
	for cur != nil && cur.Name == "" && cur.OfType != nil {
		cur = cur.OfType
	}
	if cur != nil {
		ft.Base = cur.Name
	}
	return ft
}
