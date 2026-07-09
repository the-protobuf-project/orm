package ir

import (
	"testing"

	"github.com/the-protobuf-project/orm/plugin/factory/source/graphql/dialect"
)

// A native mutation (procedure) returns its own projection rows, so resourceOf
// buckets it under the projection type. rehomeProcedures must fold it back onto
// the table family its name addresses, where method naming then yields e.g.
// Guests.DeleteByBookingId instead of a bogus one-op resource package.
func TestGroupResourcesRehomesProcedures(t *testing.T) {
	s := &Schema{
		Objects: map[string]*Object{
			"IdentityGuests": {Name: "IdentityGuests", Fields: []Field{{Name: "id", Type: FieldType{Base: "String"}}}},
			"DeleteIdentityGuestsByBookingId": {Name: "DeleteIdentityGuestsByBookingId", Fields: []Field{
				{Name: "id", Type: FieldType{Base: "String"}},
			}},
			"DeleteIdentityGuestsByBookingIdResponse": {Name: "DeleteIdentityGuestsByBookingIdResponse", Fields: []Field{
				{Name: "affectedRows", Type: FieldType{Base: "Int32"}},
				{Name: "returning", Type: FieldType{Base: "DeleteIdentityGuestsByBookingId", List: true}},
			}},
		},
		Queries: []*Operation{
			{Name: "identityGuests", Kind: "query", Return: FieldType{Base: "IdentityGuests", List: true}},
		},
		Mutations: []*Operation{
			{Name: "deleteIdentityGuestsByBookingId", Kind: "mutation", Return: FieldType{Base: "DeleteIdentityGuestsByBookingIdResponse"}},
		},
	}

	resources := groupResources(s, dialect.Default())
	if len(resources) != 1 {
		names := make([]string, 0, len(resources))
		for _, r := range resources {
			names = append(names, r.Name)
		}
		t.Fatalf("want the procedure folded into 1 resource, got %d: %v", len(resources), names)
	}
	r := resources[0]
	if r.Name != "IdentityGuests" || len(r.Queries) != 1 || len(r.Mutations) != 1 {
		t.Fatalf("procedure not re-homed onto the table family: %+v", r)
	}
	if r.Mutations[0].Name != "deleteIdentityGuestsByBookingId" {
		t.Fatalf("wrong mutation re-homed: %q", r.Mutations[0].Name)
	}
}

// A mutation-only family whose name matches no table family must keep its own
// resource (today's behavior for standalone procedures).
func TestGroupResourcesKeepsUnmatchedProcedures(t *testing.T) {
	s := &Schema{
		Objects: map[string]*Object{
			"RunNightlyAuditResponse": {Name: "RunNightlyAuditResponse", Fields: []Field{
				{Name: "affectedRows", Type: FieldType{Base: "Int32"}},
			}},
		},
		Mutations: []*Operation{
			{Name: "runNightlyAudit", Kind: "mutation", Return: FieldType{Base: "RunNightlyAuditResponse"}},
		},
	}
	resources := groupResources(s, dialect.Default())
	if len(resources) != 1 || resources[0].Name != "RunNightlyAuditResponse" {
		t.Fatalf("unmatched procedure should keep its own resource, got %+v", resources)
	}
}
