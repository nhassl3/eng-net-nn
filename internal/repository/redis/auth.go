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
	key := profileKey
	switch {
	case params.Username != nil && *params.Username != "":
		key += *params.Username
	case params.Email != nil && *params.Email != "":
		key += *params.Email
	case params.UUID != nil && *params.UUID != "":
		key += *params.UUID
	default:
		return nil, domain.ErrRedisNotFound
	}

	user := &domain.User{}
	if err := r.redis.Get(ctx, key).Scan(user); err != nil {
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

	if user.Username != "" {
		_ = r.redis.Set(ctx, profileKey+user.Username, user, r.userProfileTTL).Err()
	}
	if user.Email != "" {
		_ = r.redis.Set(ctx, profileKey+user.Email, user, r.userProfileTTL).Err()
	}

	return nil
}
