// Package ir holds the normalized intermediate representation of a GraphQL schema.
//
// It converts the raw introspection model (see the introspect package) into a
// language-agnostic shape that code generators consume: objects, enums, inputs,
// scalars, and root operations grouped per resource. Generators never re-walk the
// introspection wrappers; they read flat FieldType values and resource groupings.
package ir

// Schema is the normalized schema. Maps are keyed by GraphQL type name.
type Schema struct {
	Objects       map[string]*Object
	Enums         map[string]*Enum
	Inputs        map[string]*Input
	Scalars       map[string]bool
	Queries       []*Operation
	Mutations     []*Operation
	Subscriptions []*Operation
	Resources     []*Resource
}

// FieldType is a fully resolved type reference with the NON_NULL/LIST wrappers
// flattened. Base is the underlying named type (e.g. "BookingContacts").
type FieldType struct {
	Base        string // named base type
	List        bool   // true if the type is a list
	NonNull     bool   // true if the outer type is non-null
	ElemNonNull bool   // true if list elements are non-null
}

// Object is an OBJECT type with its fields.
type Object struct {
	Name        string
	Description string
	Fields      []Field
}

// Field is a field on an object or an input object, with optional arguments.
type Field struct {
	Name        string
	Description string
	Type        FieldType
	Args        []Arg
}

// Arg is an argument to a field or operation.
type Arg struct {
	Name string
	Type FieldType
}

// Enum is an ENUM type and its values.
type Enum struct {
	Name        string
	Description string
	Values      []string
}

// Input is an INPUT_OBJECT type and its fields.
type Input struct {
	Name        string
	Description string
	Fields      []Field
}

// Operation is a root query/mutation field.
type Operation struct {
	Name   string    // GraphQL field name (e.g. "bookingContactsById")
	Kind   string    // "query" or "mutation"
	Args   []Arg     // operation arguments
	Return FieldType // return type
}

// Resource groups the operations that act on a single row object (e.g. all
// operations for "BookingContacts"). Used to organize generated files.
type Resource struct {
	Name          string // row object type name
	Queries       []*Operation
	Mutations     []*Operation
	Subscriptions []*Operation
}

// IsScalarOrEnum reports whether base names a scalar or enum (a leaf type that does
// not require a sub-selection).
func (s *Schema) IsScalarOrEnum(base string) bool {
	if s.Scalars[base] {
		return true
	}
	_, isEnum := s.Enums[base]
	return isEnum
}
