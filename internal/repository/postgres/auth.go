package postgres

import (
	"context"
	"fmt"
	"strings"

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
		Username:       params.Username,
		FullName:       stringToNullable(params.FullName),
		Email:          params.Email,
		HashedPassword: params.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("auth_repo.CreateUser: failed to create user: %w", err)
	}
	return new(mapUser(user)), nil
}

// GetUserForLogin fetches the user and their stored password hash for login verification.
func (r *AuthRepo) GetUserForLogin(ctx context.Context, in *domain.SignInInput) (*domain.User, string, error) {
	if in.ID == "" && in.Username == "" && in.Email == "" {
		return nil, "", fmt.Errorf("auth_repo.GetUserForLogin: %w", domain.ErrInvalidParam)
	}

	user, err := r.db.GetUser(ctx, db.GetUserParams{
		Username: stringToNullable(in.Username),
		Email:    stringToNullable(in.Email),
		ID:       uuidPtr2Nullable(strings.ToLower(in.ID)),
	})
	if err != nil {
		return nil, "", fmt.Errorf("auth_repo.GetUserForLogin: %w", err)
	}
	return new(mapUser(user)), user.HashedPassword, nil
}

func (r *AuthRepo) GetMe(ctx context.Context, params domain.GetMeParams) (*domain.User, error) {
	if (params.UUID == nil || *params.UUID == "") &&
		(params.Email == nil || *params.Email == "") &&
		(params.Username == nil || *params.Username == "") {
		return nil, fmt.Errorf("auth_repo.GetMe: %w", domain.ErrInvalidParam)
	}

	user, err := r.db.GetUser(ctx, db.GetUserParams{
		ID:       nUUIDPtr2Nullable(params.UUID),
		Email:    stringPtrToNullable(params.Email),
		Username: stringPtrToNullable(params.Username),
	})
	if err != nil {
		return nil, fmt.Errorf("auth_repo.GetMe: failed to get user: %w", err)
	}
	return new(mapUser(user)), nil
}

func mapUser(user db.User) domain.User {
	return domain.User{
		UUID:      uuid2String(user.ID),
		Role:      "",
		Username:  user.Username,
		FullName:  user.FullName.String,
		Email:     user.Email,
		CreatedAt: pgTimeTZ(user.CreatedAt),
		UpdatedAt: pgTimeTZ(user.UpdatedAt),
	}
}
