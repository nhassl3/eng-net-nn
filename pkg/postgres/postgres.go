package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/nhassl3/IpBuild-backend/pkg/logger"
)

// tracelogAdapter routes pgx's own query/connection logging through the
// application's structured logger instead of stdlib log. Query arguments
// are not logged: they can carry passwords, tokens or other PII.
type tracelogAdapter struct {
	log logger.Logger
}

func (a tracelogAdapter) Log(_ context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	fields := make([]logger.Field, 0, len(data))
	for k, v := range data {
		if k == "args" || k == "sql" {
			// Skip query text/args: may contain sensitive values, and the
			// operation name/duration below is enough to diagnose slow
			// queries or driver errors without leaking data.
			continue
		}
		fields = append(fields, logger.Any(k, v))
	}

	switch level {
	case tracelog.LogLevelTrace, tracelog.LogLevelDebug:
		a.log.Debug(msg, fields...)
	case tracelog.LogLevelInfo:
		a.log.Debug(msg, fields...) // pgx logs every query at "info"; too noisy for our info level
	case tracelog.LogLevelWarn:
		a.log.Warn(msg, fields...)
	case tracelog.LogLevelError:
		a.log.Error(msg, fields...)
	default:
		a.log.Info(msg, fields...)
	}
}

const (
	defaultMaxConns          = 25
	defaultMinConns          = 5
	defaultMaxConnLifeTime   = time.Hour
	defaultMaxConnIdleTime   = 30 * time.Minute
	defaultHealthCheckPeriod = time.Minute
	defaultConnectTimeout    = 5 * time.Second
)

type Options struct {
	MaxConns          int32
	MinConns          int32
	MaxConnLifeTime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	ConnectTimeout    time.Duration
}

func NewPool(ctx context.Context, dsn string, log logger.Logger, opts ...Options) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: parse config: %w", err)
	}

	cfg.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   tracelogAdapter{log: log},
		LogLevel: tracelog.LogLevelWarn,
	}

	o := Options{
		MaxConns:          defaultMaxConns,
		MinConns:          defaultMinConns,
		MaxConnLifeTime:   defaultMaxConnLifeTime,
		MaxConnIdleTime:   defaultMaxConnIdleTime,
		HealthCheckPeriod: defaultHealthCheckPeriod,
		ConnectTimeout:    defaultConnectTimeout,
	}
	if len(opts) > 0 {
		o = opts[0]
	}

	cfg.MaxConns = o.MaxConns
	cfg.MinConns = o.MinConns
	cfg.MaxConnLifetime = o.MaxConnLifeTime
	cfg.MaxConnIdleTime = o.MaxConnIdleTime
	cfg.HealthCheckPeriod = o.HealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("postgres: create pool: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}

	return pool, nil
}

func DSN(host string, port int, user, password, dbName, sslMode string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", user, password, host, port, dbName, sslMode)
}
