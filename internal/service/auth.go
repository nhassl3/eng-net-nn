package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/nhassl3/IpBuild-backend/internal/domain"
	"github.com/nhassl3/IpBuild-backend/internal/repository/postgres"
	"github.com/nhassl3/IpBuild-backend/internal/repository/redis"
	"github.com/nhassl3/IpBuild-backend/pkg/auth"
	"github.com/nhassl3/IpBuild-backend/pkg/hash"
)

type AuthService struct {
	repo         postgres.Authorization
	adminRepo    postgres.Admin
	redisRepo    redis.AuthRedis
	accessMaker  auth.TokenManager
	refreshMaker auth.TokenManager
	blacklist    auth.TokenBlacklist
}

func NewAuthService(
	repo postgres.Authorization,
	adminRepo postgres.Admin,
	redisRepo redis.AuthRedis,
	accessMaker auth.TokenManager,
	refreshMaker auth.TokenManager,
	blacklist auth.TokenBlacklist,
) *AuthService {
	return &AuthService{
		repo:         repo,
		adminRepo:    adminRepo,
		redisRepo:    redisRepo,
		accessMaker:  accessMaker,
		refreshMaker: refreshMaker,
		blacklist:    blacklist,
	}
}

func (s *AuthService) CreateUser(ctx context.Context, input *domain.CreateUserInput) (*domain.User, *domain.TokenPair, error) {
	hashedPwd, err := hash.CreateHashPassword(input.Password)
	if err != nil {
		return nil, nil, fmt.Errorf("auth_service.CreateUser: hash password: %w", err)
	}

	dbInput := domain.CreateUserInput{
		Username: input.Username,
		FullName: input.FullName,
		Email:    input.Email,
		Password: hashedPwd,
	}

	user, err := s.repo.CreateUser(ctx, dbInput)
	if err != nil {
		if isDuplicateError(err) {
			return nil, nil, domain.ErrUserAlreadyExists
		}
		return nil, nil, fmt.Errorf("auth_service.CreateUser: %w", err)
	}

	tokenPair, err := s.GenerateToken(ctx, user)
	if err != nil {
		return nil, nil, fmt.Errorf("auth_service.CreateUser: generate tokens: %w", err)
	}

	_ = s.redisRepo.SetProfile(ctx, user)

	return user, tokenPair, nil
}

func (s *AuthService) SignIn(ctx context.Context, req *domain.SignInInput) (*domain.User, error) {
	user, storedHash, err := s.repo.GetUserForLogin(ctx, req)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	ok, err := hash.VerifyPassword(req.Password, storedHash)
	if err != nil || !ok {
		return nil, domain.ErrInvalidCredentials
	}

	isAdmin, err := s.adminRepo.IsAdmin(ctx, user.UUID)
	if err == nil && isAdmin {
		user.Role = "admin"
	}

	_ = s.redisRepo.SetProfile(ctx, user)

	return user, nil
}

func (s *AuthService) GenerateToken(ctx context.Context, user *domain.User) (*domain.TokenPair, error) {
	accessToken, err := s.accessMaker.CreateToken(user.Username, user.UUID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("auth_service.GenerateToken: access: %w", err)
	}

	refreshToken, _, err := s.refreshMaker.CreateRefreshToken(user.Username, user.UUID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("auth_service.GenerateToken: refresh: %w", err)
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) ParseToken(ctx context.Context, token string) (*domain.User, error) {
	payload, err := s.accessMaker.VerifyToken(token)
	if err != nil {
		return nil, err
	}

	return &domain.User{
		UUID:     payload.UID,
		Username: payload.Username,
		Role:     payload.Role,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	payload, err := s.refreshMaker.VerifyToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("auth_service.RefreshToken: %w", err)
	}

	user := &domain.User{
		UUID:     payload.UID,
		Username: payload.Username,
		Role:     payload.Role,
	}

	return s.GenerateToken(ctx, user)
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	payload, err := s.accessMaker.VerifyToken(token)
	if err != nil {
		return fmt.Errorf("auth_service.Logout: %w", err)
	}

	if err := s.blacklist.Blacklist(ctx, payload.JTI, payload.ExpiredAt); err != nil {
		return fmt.Errorf("auth_service.Logout: blacklist: %w", err)
	}

	return nil
}

func (s *AuthService) GetMe(ctx context.Context, uuid string) (*domain.User, error) {
	user, err := s.redisRepo.Profile(ctx, domain.GetMeParams{
		UUID: &uuid,
	})
	if err != nil {
		user, err = s.repo.GetMe(ctx, domain.GetMeParams{UUID: &uuid})
		if err != nil {
			return nil, fmt.Errorf("auth_service.GetMe: %w", err)
		}
		if errors.Is(err, domain.ErrRedisNotFound) {
			_ = s.redisRepo.SetProfile(ctx, user)
		}
	}
	return user, nil
}

func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, domain.ErrUserAlreadyExists) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "23505")
}
