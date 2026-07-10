{{.Header}}

package filterx

import (
	"context"

	pulse "{{.PulseImport}}"
)

// PulseObserver adapts a pulse-go runtime to the Observer interface the list
// engines report through: one span per query and a debug log per rejected
// filter/order input. Wire it when building an engine:
//
//	filterx.Gorm[Model](spec).Observe(filterx.PulseObserver(shared.Pulse))
func PulseObserver(p *pulse.Pulse) Observer {
	if p == nil {
		return NopObserver{}
	}
	return pulseObserver{p: p}
}

type pulseObserver struct{ p *pulse.Pulse }

// Span wraps fn in a pulse trace span.
func (o pulseObserver) Span(ctx context.Context, name string, fn func(context.Context) error) error {
	return o.p.Tracing.Trace(ctx, name, nil, func(ctx context.Context, _ *pulse.Span) error {
		return fn(ctx)
	})
}

// Debug logs a non-fatal engine event.
func (o pulseObserver) Debug(msg string, kv map[string]any) {
	o.p.Logger.Debug(msg, kv)
}
