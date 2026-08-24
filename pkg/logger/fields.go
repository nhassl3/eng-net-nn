package logger

import (
	"time"

	"go.uber.org/zap"
)

// Err builds an "error" field. It is nil-safe: a nil error logs as null
// rather than panicking, so call sites don't need to guard it.
func Err(err error) Field {
	if err == nil {
		return zap.String("error", "")
	}
	return zap.String("error", err.Error())
}

// RequestID builds a "request_id" field.
func RequestID(id string) Field {
	return zap.String("request_id", id)
}

// UserID builds a "user_id" field.
func UserID(id string) Field {
	return zap.String("user_id", id)
}

// Op builds an "op" field identifying the operation, e.g. "vacancy.Create".
func Op(op string) Field {
	return zap.String("op", op)
}

// Duration builds a "duration_ms" field.
func Duration(d time.Duration) Field {
	return zap.Duration("duration_ms", d)
}

// String, Int, Bool, Any are thin re-exports of the zap constructors so most
// call sites only need to import pkg/logger.

func String(key, value string) Field {
	if IsSensitiveKey(key) {
		return Masked(key, value)
	}
	return zap.String(key, value)
}
func Int(key string, value int) Field   { return zap.Int(key, value) }
func Bool(key string, value bool) Field { return zap.Bool(key, value) }
func Any(key string, value any) Field   { return zap.Any(key, value) }
