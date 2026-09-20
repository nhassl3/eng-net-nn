package redis

import (
	"context"

	"github.com/nhassl3/IpBuild-backend/internal/domain"
)

type AuthRedis interface {
	Profile(ctx context.Context, params domain.GetMeParams) (*domain.User, error)
	SetProfile(ctx context.Context, user *domain.User) error
	DeleteProfile(ctx context.Context, uuid string) error
}
