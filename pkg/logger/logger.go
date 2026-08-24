// Package logger provides a thin, structured logging interface backed by
// go.uber.org/zap. It intentionally avoids colorful/console handlers — every
// environment logs JSON, suitable for aggregation in a highload production
// system. Callers depend on the Logger interface, not on zap directly, so the
// backend can be swapped or mocked without touching call sites.
package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Field is an alias for zap.Field so callers can build fields with zap.* or
// with the helpers in fields.go without importing zap themselves.
type Field = zap.Field

// Logger is the structured logger interface used across the application.
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)

	// With returns a child logger that always includes the given fields.
	With(fields ...Field) Logger
	// Named returns a child logger scoped to the given component name
	// (e.g. "service.vacancy", "repo.postgres"). Names are dot-joined.
	Named(name string) Logger

	// Sync flushes any buffered log entries. Call it before process exit.
	Sync() error
}

// Config controls how New builds the logger.
type Config struct {
	// Level is one of debug|info|warn|error. Empty defaults to "info".
	Level string
	// AddCaller adds the "caller" field (file:line of the log call).
	AddCaller bool
	// Stacktrace is the minimum level at which a stacktrace is attached.
	// Empty defaults to "error".
	Stacktrace string
	// Env is a free-form deployment environment, added as a constant field.
	Env string
	// Service is the service name, added as a constant field.
	Service string
	// Version is the build version, added as a constant field.
	Version string
}

type zapLogger struct {
	l *zap.Logger
}

// New builds a JSON-only, structured Logger. There is no console/colorful
// handler: every environment (local included) logs machine-parsable JSON, so
// logs behave the same wherever they end up (terminal, file, log shipper).
func New(cfg Config) (Logger, error) {
	level := parseLevel(cfg.Level)
	stacktraceLevel := parseLevel(cfg.Stacktrace)
	if cfg.Stacktrace == "" {
		stacktraceLevel = zapcore.ErrorLevel
	}

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		zap.NewAtomicLevelAt(level),
	)

	opts := []zap.Option{
		zap.AddStacktrace(stacktraceLevel),
	}
	if cfg.AddCaller {
		opts = append(opts, zap.AddCaller())
	}

	fields := make([]Field, 0, 3)
	if cfg.Service != "" {
		fields = append(fields, zap.String("service", cfg.Service))
	}
	if cfg.Env != "" {
		fields = append(fields, zap.String("env", cfg.Env))
	}
	if cfg.Version != "" {
		fields = append(fields, zap.String("version", cfg.Version))
	}
	if len(fields) > 0 {
		opts = append(opts, zap.Fields(fields...))
	}

	return &zapLogger{l: zap.New(core, opts...)}, nil
}

// Nop returns a Logger that discards everything. Use it in tests and as the
// safe fallback when no logger is available.
func Nop() Logger {
	return &zapLogger{l: zap.NewNop()}
}

func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "info", "":
		return zapcore.InfoLevel
	default:
		return zapcore.InfoLevel
	}
}

func (z *zapLogger) Debug(msg string, fields ...Field) { z.l.Debug(msg, fields...) }
func (z *zapLogger) Info(msg string, fields ...Field)  { z.l.Info(msg, fields...) }
func (z *zapLogger) Warn(msg string, fields ...Field)  { z.l.Warn(msg, fields...) }
func (z *zapLogger) Error(msg string, fields ...Field) { z.l.Error(msg, fields...) }

func (z *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{l: z.l.With(fields...)}
}

func (z *zapLogger) Named(name string) Logger {
	return &zapLogger{l: z.l.Named(name)}
}

func (z *zapLogger) Sync() error {
	return z.l.Sync()
}
