package logger

import (
	"strings"

	"go.uber.org/zap"
)

// mask is the placeholder written in place of a redacted value.
const mask = "***"

// Masked builds a field whose value is always logged as "***", regardless of
// the underlying value. Use it for passwords, tokens, API keys, PASETO
// secrets, etc. — anything that must never reach a log sink.
func Masked(key, _ string) Field {
	return zap.String(key, mask)
}

// Email builds a field with a partially masked email address, e.g.
// "jo***@gmail.com". Useful for correlating log lines by user without
// leaking the full address into log storage.
func Email(key, addr string) Field {
	return zap.String(key, maskEmail(addr))
}

func maskEmail(addr string) string {
	at := strings.IndexByte(addr, '@')
	if at <= 0 {
		return mask
	}
	local, domain := addr[:at], addr[at:]
	if len(local) <= 2 {
		return local[:1] + mask + domain
	}
	return local[:2] + mask + domain
}

// SensitiveKeys lists field/attribute names that must never be logged as
// plain text. Callers building generic key/value logs (e.g. dumping a
// struct or a query's named params) should route any key from this set
// through Masked instead of a plain string field.
var SensitiveKeys = map[string]struct{}{
	"password":      {},
	"token":         {},
	"access_token":  {},
	"refresh_token": {},
	"authorization": {},
	"paseto":        {},
	"paseto_key":    {},
	"secret_key":    {},
	"api_key":       {},
	"dsn":           {},
}

// IsSensitiveKey reports whether key names a value that must be masked
// before logging, matching case-insensitively.
func IsSensitiveKey(key string) bool {
	_, ok := SensitiveKeys[strings.ToLower(key)]
	return ok
}
