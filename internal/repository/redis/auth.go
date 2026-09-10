package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nhassl3/IpBuild-backend/internal/domain"
	"github.com/redis/go-redis/v9"
)

const profileKey = "user:profile:"

type AuthRedisRepository struct {
	redis          *redis.Client
	userProfileTTL time.Duration
}

func NewAuthRedisRepository(redis *redis.Client, userProfileTTL time.Duration) *AuthRedisRepository {
	return &AuthRedisRepository{
		redis:          redis,
		userProfileTTL: userProfileTTL,
	}
}

func (r *AuthRedisRepository) Profile(ctx context.Context, params domain.GetMeParams) (*domain.User, error) {
	if params.UUID == nil || *params.UUID == "" {
		return nil, domain.ErrRedisNotFound
	}

	var user *domain.User
	if err := r.redis.Get(ctx, profileKey+*params.UUID).Scan(user); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrRedisNotFound
		}
		return nil, fmt.Errorf("AuthRedis.Profile: scan: %w", err)
	}

	return user, nil
}

func (r *AuthRedisRepository) SetProfile(ctx context.Context, user *domain.User) error {
	if user == nil {
		return nil
	}

	key := profileKey + user.UUID
	if err := r.redis.Set(ctx, key, user, r.userProfileTTL).Err(); err != nil {
		return fmt.Errorf("AuthRedis.SetProfile: %w", err)
	}

	return nil
}
