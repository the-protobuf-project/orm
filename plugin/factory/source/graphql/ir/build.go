package ir

import (
	"strings"

	"github.com/the-protobuf-project/orm/plugin/factory/source/graphql/dialect"
	"github.com/the-protobuf-project/orm/plugin/factory/source/graphql/introspect"
)

// Build normalizes an introspection Schema into the IR, classifying types, resolving
// type references, and grouping root operations per resource. The dialect supplies
// the engine conventions used to recognize mutation/aggregate wrappers and re-home
// procedure-style mutations.
func Build(in *introspect.Schema, d dialect.Dialect) *Schema {
	s := &Schema{
		Objects: map[string]*Object{},
		Enums:   map[string]*Enum{},
		Inputs:  map[string]*Input{},
		Scalars: map[string]bool{},
	}

	queryRoot := nameOf(in.QueryType)
	mutationRoot := nameOf(in.MutationType)
	subscriptionRoot := nameOf(in.SubscriptionType)

	for i := range in.Types {
		t := &in.Types[i]
		if strings.HasPrefix(t.Name, "__") {
			continue // skip introspection meta-types
		}
		switch t.Kind {
		case "SCALAR":
			s.Scalars[t.Name] = true
		case "ENUM":
			s.Enums[t.Name] = buildEnum(t)
		case "INPUT_OBJECT":
			s.Inputs[t.Name] = buildInput(t)
		case "OBJECT":
			// Root operation containers are handled separately, not as models.
			if t.Name == queryRoot || t.Name == mutationRoot || t.Name == subscriptionRoot {
				continue
			}
			s.Objects[t.Name] = buildObject(t)
		}
	}

	for i := range in.Types {
		t := &in.Types[i]
		switch t.Name {
		case queryRoot:
			s.Queries = buildOperations(t, "query")
		case mutationRoot:
			s.Mutations = buildOperations(t, "mutation")
		case subscriptionRoot:
			s.Subscriptions = buildOperations(t, "subscription")
		}
	}

	s.Resources = groupResources(s, d)
	return s
}

func nameOf(t *introspect.TypeName) string {
	if t == nil {
		return ""
	}
	return t.Name
}

func buildEnum(t *introspect.FullType) *Enum {
	e := &Enum{Name: t.Name, Description: t.Description}
	for _, v := range t.EnumValues {
		e.Values = append(e.Values, v.Name)
	}
	return e
}

func buildInput(t *introspect.FullType) *Input {
	in := &Input{Name: t.Name, Description: t.Description}
	for _, f := range t.InputFields {
		in.Fields = append(in.Fields, Field{Name: f.Name, Description: f.Description, Type: resolve(f.Type)})
	}
	return in
}

func buildObject(t *introspect.FullType) *Object {
	o := &Object{Name: t.Name, Description: t.Description}
	for _, f := range t.Fields {
		o.Fields = append(o.Fields, Field{Name: f.Name, Description: f.Description, Type: resolve(f.Type)})
	}
	return o
}

func buildOperations(t *introspect.FullType, kind string) []*Operation {
	ops := make([]*Operation, 0, len(t.Fields))
	for _, f := range t.Fields {
		op := &Operation{Name: f.Name, Kind: kind, Return: resolve(f.Type)}
		for _, a := range f.Args {
			op.Args = append(op.Args, Arg{Name: a.Name, Type: resolve(a.Type)})
		}
		ops = append(ops, op)
	}
	return ops
}
