package generator

// diag.go collects non-fatal schema problems found while building the IR and
// decides how to surface each one — warning or hard error — based on the
// per-rule `strict` option. Every problem carries a rule key ("ref",
// "collision", "index", "lint") so a project can fail CI on the categories it
// cares about while tolerating the rest.

import (
	"fmt"
	"os"
	"strings"
)

// diag is one recorded problem and the rule category it belongs to.
type diag struct {
	rule string
	msg  string
}

// diagnostics accumulates schema problems that are recoverable but suspicious —
// an unresolved resource_reference, a colliding model name, an index naming a
// column that does not exist. resolve decides per rule whether each is a warning
// (printed to stderr; codegen proceeds with a best-effort fallback) or a hard
// error that fails the build.
type diagnostics struct {
	items []diag
}

// warnf records one problem under the given rule.
func (d *diagnostics) warnf(rule, format string, args ...any) {
	d.items = append(d.items, diag{rule: rule, msg: fmt.Sprintf(format, args...)})
}

// resolve surfaces every recorded problem at the severity the strict spec
// assigns its rule. Problems whose rule is "error" are aggregated into one hard
// error; the rest print to stderr and generation continues.
func (d *diagnostics) resolve(spec string) error {
	if len(d.items) == 0 {
		return nil
	}
	sev := parseStrict(spec)
	var errs []string
	for _, it := range d.items {
		line := "[" + it.rule + "] " + it.msg
		if sev.isError(it.rule) {
			errs = append(errs, line)
		} else {
			fmt.Fprintln(os.Stderr, "protorm: warning: "+line)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("protorm: strict: %d schema problem(s):\n  - %s",
			len(errs), strings.Join(errs, "\n  - "))
	}
	return nil
}

// severity decides whether a given rule is a hard error. def is the fallback for
// rules without an explicit entry.
type severity struct {
	def   bool
	rules map[string]bool
}

func (s severity) isError(rule string) bool {
	if v, ok := s.rules[rule]; ok {
		return v
	}
	return s.def
}

// parseStrict reads the strict spec. Accepted forms (comma-separated):
//
//	""            no rule is an error (every problem warns)  [default]
//	"true"        every rule is an error
//	"false"       every rule warns
//	"ref:error,collision:warn,index:error"   per-rule severity
//	"*:error" / "default:error"              set the fallback for unlisted rules
//
// A bare "true"/"false" sets the fallback; per-rule entries override it.
func parseStrict(spec string) severity {
	s := severity{rules: map[string]bool{}}
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		switch part {
		case "":
			continue
		case "true":
			s.def = true
			continue
		case "false":
			s.def = false
			continue
		}
		key, val, ok := strings.Cut(part, ":")
		if !ok {
			continue
		}
		isErr := strings.TrimSpace(val) == "error"
		if key := strings.TrimSpace(key); key == "*" || key == "default" {
			s.def = isErr
		} else {
			s.rules[key] = isErr
		}
	}
	return s
}
