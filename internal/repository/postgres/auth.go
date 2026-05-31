package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/nhassl3/IpBuild-backend/internal/db"
	"github.com/nhassl3/IpBuild-backend/internal/domain"
)

type AuthRepo struct {
	db *db.Store
}

func NewAuthRepo(db *db.Store) *AuthRepo {
	return &AuthRepo{
		db: db,
	}
}

func (r *AuthRepo) CreateUser(ctx context.Context, params domain.CreateUserInput) (*domain.User, error) {
	user, err := r.db.CreateUser(ctx, db.CreateUserParams{
		Username:       stringToNullable(params.Username),
		FullName:       stringToNullable(params.FullName),
		Email:          params.Email,
		HashedPassword: params.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("auth_repo.CreateUser: failed to create user: %w", err)
	}
	domainUser := mapUser(user)
	return &domainUser, nil
}

// GetUserForLogin fetches the user and their stored password hash for login verification.
func (r *AuthRepo) GetUserForLogin(ctx context.Context, username string) (*domain.User, string, error) {
	user, err := r.db.GetUser(ctx, db.GetUserParams{
		Username: stringToNullable(username),
	})
	if err != nil {
		return nil, "", fmt.Errorf("auth_repo.GetUserForLogin: %w", err)
	}
	domainUser := mapUser(user)
	return &domainUser, user.HashedPassword, nil
}

func (r *AuthRepo) GetMe(ctx context.Context, params domain.GetMeParams) (*domain.User, error) {
	user, err := r.db.GetUser(ctx, db.GetUserParams{
		ID:       nUUIDPtr2Nullable(params.UUID),
		Email:    usernamePtrToNullable(params.Email),
		Username: usernamePtrToNullable(params.Username),
	})
	if err != nil {
		return nil, fmt.Errorf("auth_repo.GetMe: failed to get user: %w", err)
	}
	domainUser := mapUser(user)
	return &domainUser, nil
}

func mapUser(user db.User) domain.User {
	return domain.User{
		UUID:      uuid2String(user.ID),
		Username:  user.Username.String,
		FullName:  user.FullName.String,
		Email:     user.Email,
		CreatedAt: pgTimeTZ(user.CreatedAt, time.UTC),
		UpdatedAt: pgTimeTZ(user.UpdatedAt, time.UTC),
	}
}
