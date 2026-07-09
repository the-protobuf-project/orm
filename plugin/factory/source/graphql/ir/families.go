package ir

import (
	"sort"
	"strings"

	"github.com/the-protobuf-project/orm/plugin/factory/source/graphql/dialect"
)

// groupResources buckets root operations by the row object they act on, so the
// generator can emit one set of files per resource. The resource is derived from
// each operation's return type (see resourceOf).
func groupResources(s *Schema, d dialect.Dialect) []*Resource {
	byName := map[string]*Resource{}
	var order []string

	add := func(name string, op *Operation) {
		r, ok := byName[name]
		if !ok {
			r = &Resource{Name: name}
			byName[name] = r
			order = append(order, name)
		}
		switch op.Kind {
		case "mutation":
			r.Mutations = append(r.Mutations, op)
		case "subscription":
			r.Subscriptions = append(r.Subscriptions, op)
		default:
			r.Queries = append(r.Queries, op)
		}
	}

	for _, op := range s.Queries {
		add(resourceOf(s, op, d), op)
	}
	for _, op := range s.Mutations {
		add(resourceOf(s, op, d), op)
	}
	for _, op := range s.Subscriptions {
		add(resourceOf(s, op, d), op)
	}

	rehomed := rehomeProcedures(byName, d)

	sort.Strings(order)
	resources := make([]*Resource, 0, len(order))
	for _, name := range order {
		if rehomed[name] {
			continue
		}
		resources = append(resources, byName[name])
	}
	return resources
}

// rehomeProcedures folds procedure-style mutation families into the table family
// their name addresses. A native mutation (e.g. delete_identity_guests_by_booking_id)
// returns its own projection rows, so resourceOf buckets it under the projection
// type ("DeleteIdentityGuestsByBookingId") rather than the table it acts on
// ("IdentityGuests") — which would generate a bogus one-op resource package. Any
// family that consists only of mutations and is named <Verb><TableFamily><Rest>
// has its ops moved onto that table family (the longest-named match wins), where
// the method-name derivation then yields e.g. Guests.DeleteByBookingId. Returns
// the set of family names that were emptied and must be dropped.
func rehomeProcedures(byName map[string]*Resource, d dialect.Dialect) map[string]bool {
	verbs := make([]string, 0, len(d.MutationVerbs()))
	for _, v := range d.MutationVerbs() {
		verbs = append(verbs, v.NamePrefix)
	}
	names := make([]string, 0, len(byName))
	for name := range byName {
		names = append(names, name)
	}
	sort.Strings(names) // deterministic re-homing (and method) order

	rehomed := map[string]bool{}
	for _, name := range names {
		r := byName[name]
		if len(r.Mutations) == 0 || len(r.Queries) > 0 || len(r.Subscriptions) > 0 {
			continue
		}
		base := ""
		for _, v := range verbs {
			if strings.HasPrefix(name, v) && len(name) > len(v) {
				base = name[len(v):]
				break
			}
		}
		if base == "" {
			continue
		}
		target := ""
		for _, tname := range names {
			if tname == name || len(byName[tname].Queries) == 0 {
				continue // only real table families (they carry query roots) qualify
			}
			if strings.HasPrefix(base, tname) && len(tname) > len(target) {
				target = tname
			}
		}
		if target == "" {
			continue
		}
		byName[target].Mutations = append(byName[target].Mutations, r.Mutations...)
		rehomed[name] = true
	}
	return rehomed
}

// resourceOf determines the row-object name an operation belongs to. It unwraps the
// return type and, for mutation response wrappers (which expose a "returning" list of
// rows) and aggregate wrappers, maps back to the underlying row object.
func resourceOf(s *Schema, op *Operation, d dialect.Dialect) string {
	base := op.Return.Base
	obj, ok := s.Objects[base]
	if !ok {
		if base == "" {
			return "Root"
		}
		return base
	}

	// Mutation wrappers: {affectedRows, returning: [Row!]!} -> Row.
	for _, f := range obj.Fields {
		if f.Name == d.ReturningField() && f.Type.List {
			if _, isObj := s.Objects[f.Type.Base]; isObj {
				return f.Type.Base
			}
		}
	}

	// Aggregate wrappers: XAggExp / XAggregate -> X when the row object exists.
	for _, suffix := range d.AggregateSuffixes() {
		if trimmed := strings.TrimSuffix(base, suffix); trimmed != base {
			if _, isObj := s.Objects[trimmed]; isObj {
				return trimmed
			}
		}
	}
	return base
}
