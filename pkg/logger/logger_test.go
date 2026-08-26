package logger

import (
	"context"
	"errors"
	"testing"
)

func TestErrNilSafe(t *testing.T) {
	f := Err(nil)
	if f.String != "" {
		t.Fatalf("expected empty string for nil error, got %q", f.String)
	}
}

func TestRedactMasksSensitiveValue(t *testing.T) {
	f := Masked("password", "hunter2")
	if f.String != "***" {
		t.Fatalf("expected masked value, got %q", f.String)
	}
}

func TestEmailMasking(t *testing.T) {
	got := maskEmail("john.doe@example.com")
	if got != "jo***@example.com" {
		t.Fatalf("unexpected masked email: %q", got)
	}
	if maskEmail("not-an-email") != "***" {
		t.Fatalf("expected full mask for invalid email")
	}
}

func TestFromContextFallsBackToNop(t *testing.T) {
	l := From(context.Background())
	if l == nil {
		t.Fatal("From must never return nil")
	}
	// Should not panic.
	l.Info("noop", Err(errors.New("boom")))
}

func TestInjectAndFromRoundtrip(t *testing.T) {
	base, err := New(Config{Level: "debug"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := Inject(context.Background(), base)
	got := From(ctx)
	if got == nil {
		t.Fatal("expected non-nil logger from context")
	}
}
