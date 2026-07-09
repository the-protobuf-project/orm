package dialect

// hasura is the built-in dialect for the Hasura / Hasura DDN / Grafbase /
// Prisma-backed GraphQL lineage: `_and`/`_or`/`_not` combinators, `_eq`/`_in`
// comparisons, insert/update/delete/upsert verb prefixes, `{affectedRows,
// returning}` mutation responses, and the `x-hasura-admin-secret` auth header.
// Its constants reproduce exactly what the generator previously hardcoded, so
// selecting it changes nothing about the output.
func init() { Register(hasura{}) }

// Default is the dialect used when none is configured. Hasura's lineage (Hasura
// DDN / Grafbase / Prisma-backed engines) is the common case.
func Default() Dialect { return hasura{} }

type hasura struct{}

func (hasura) Name() string { return "hasura" }

func (hasura) AuthHeader() string { return "x-hasura-admin-secret" }

func (hasura) Combinators() (and, or, not string) { return "_and", "_or", "_not" }

func (hasura) EqOperators() []string { return []string{"_eq", "_in"} }

func (hasura) SetOperands() []string { return []string{"set", "_set"} }

func (hasura) MutationVerbs() []Verb {
	return []Verb{
		{OpPrefix: "insert", NamePrefix: "Insert", Friendly: "Create"},
		{OpPrefix: "update", NamePrefix: "Update", Friendly: "Update"},
		{OpPrefix: "delete", NamePrefix: "Delete", Friendly: "Delete"},
		{OpPrefix: "upsert", NamePrefix: "Upsert", Friendly: "Upsert"},
	}
}

func (hasura) ByIdSuffix() string { return "ById" }

func (hasura) ReturningField() string { return "returning" }

func (hasura) AffectedRowsFields() []string { return []string{"affectedRows", "affected_rows"} }

func (hasura) AggregateSuffixes() []string { return []string{"AggExp", "Aggregate"} }

func (hasura) PreCheckArgs() []string { return []string{"preCheck", "pre_check"} }

func (hasura) DefaultScalars() map[string]string {
	return map[string]string{
		"ID":       "string",
		"String":   "string",
		"String1":  "string",
		"Boolean":  "bool",
		"Boolean1": "bool",
		"Int":      "int",
		"Int32":    "int32",
		"Int64":    "graphql.Int64", // engine serializes 64-bit ints as strings; flexible scalar

		"Float":      "float64",
		"Float64":    "float64",
		"Bigdecimal": "graphql.Bigdecimal", // engine returns it as string or number; flexible scalar

		"Json":        "json.RawMessage",
		"Timestamp":   "string",
		"Timestamptz": "string",

		// OrderBy (sort direction) is provided by the runtime graphql package, not generated.
		"OrderBy": "graphql.OrderBy",
	}
}
