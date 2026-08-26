package logger

import "context"

type ctxKey struct{}

// Inject returns a new context carrying the given logger.
func Inject(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// From extracts the logger stored in ctx. If none was injected it returns a
// no-op logger — never nil — so callers can always call methods on the
// result without a nil check.
func From(ctx context.Context) Logger {
	if l, ok := ctx.Value(ctxKey{}).(Logger); ok && l != nil {
		return l
	}
	return Nop()
}
