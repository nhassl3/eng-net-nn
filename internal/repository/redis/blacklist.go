package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const blacklistKey = "blacklist:"

type BlacklistRepository struct {
	redis *redis.Client
}

func NewBlacklistRepository(redis *redis.Client) *BlacklistRepository {
	return &BlacklistRepository{redis: redis}
}

func (r *BlacklistRepository) Blacklist(ctx context.Context, jti string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}
	return r.redis.Set(ctx, blacklistKey+jti, "1", ttl).Err()
}

func (r *BlacklistRepository) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	n, err := r.redis.Exists(ctx, blacklistKey+jti).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
