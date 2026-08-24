package redis

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/nhassl3/IpBuild-backend/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// slowCommandThreshold is the duration above which a redis command is
// logged as slow, regardless of whether it succeeded.
const slowCommandThreshold = 100 * time.Millisecond

func NewRedisClient(ctx context.Context, address, username, password string, db int, log logger.Logger) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Username: username,
		Password: password,
		DB:       db,
	})
	client.AddHook(loggingHook{log: log})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return client, nil
}

// loggingHook logs failed and slow redis commands. redis.Nil (key miss) is
// not an error worth logging — it's the expected outcome of a cache lookup.
type loggingHook struct {
	log logger.Logger
}

func (h loggingHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (h loggingHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmd)
		elapsed := time.Since(start)

		if err != nil && !errors.Is(err, redis.Nil) {
			h.log.Warn("redis command failed",
				logger.String("cmd", cmd.Name()), logger.Duration(elapsed), logger.Err(err))
		} else if elapsed >= slowCommandThreshold {
			h.log.Warn("slow redis command",
				logger.String("cmd", cmd.Name()), logger.Duration(elapsed))
		}
		return err
	}
}

func (h loggingHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmds)
		elapsed := time.Since(start)

		if err != nil && !errors.Is(err, redis.Nil) {
			h.log.Warn("redis pipeline failed", logger.Int("commands", len(cmds)), logger.Duration(elapsed), logger.Err(err))
		} else if elapsed >= slowCommandThreshold {
			h.log.Warn("slow redis pipeline", logger.Int("commands", len(cmds)), logger.Duration(elapsed))
		}
		return err
	}
}
