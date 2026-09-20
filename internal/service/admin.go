package service

import (
	"context"
	"fmt"

	"github.com/nhassl3/IpBuild-backend/internal/domain"
	"github.com/nhassl3/IpBuild-backend/internal/repository/postgres"
)

type AdminService struct {
	adminRepo postgres.Admin
	authRepo  postgres.Authorization
}

func NewAdminService(adminRepo postgres.Admin, authRepo postgres.Authorization) *AdminService {
	return &AdminService{
		adminRepo: adminRepo,
		authRepo:  authRepo,
	}
}

func (svc *AdminService) AddAdmin(ctx context.Context, userUID string) error {
	if err := svc.adminRepo.AddAdmin(ctx, userUID); err != nil {
		return fmt.Errorf("service.admin: AddAdmin: failed to create new admin: %w", err)
	}
	return nil
}

func (svc *AdminService) AddPartner(ctx context.Context, userUID string) error {
	if err := svc.adminRepo.AddPartner(ctx, userUID); err != nil {
		return fmt.Errorf("service.partner: AddPartner: failed to create new partner: %w", err)
	}
	return nil
}

// SearchUserByUsername looks up a user by exact username and reports their
// current role, so the admin UI can resolve a username to a uuid before
// assigning a role.
func (svc *AdminService) SearchUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	user, err := svc.authRepo.GetMe(ctx, domain.GetMeParams{Username: &username})
	if err != nil {
		return nil, fmt.Errorf("service.admin: SearchUserByUsername: %w", err)
	}

	isAdmin, err := svc.adminRepo.IsAdmin(ctx, user.UUID)
	if err != nil {
		return nil, fmt.Errorf("service.admin: SearchUserByUsername: %w", err)
	}
	isPartner, err := svc.adminRepo.IsPartner(ctx, user.UUID)
	if err != nil {
		return nil, fmt.Errorf("service.admin: SearchUserByUsername: %w", err)
	}

	switch {
	case isAdmin:
		user.Role = "admin"
	case isPartner:
		user.Role = "partner"
	default:
		user.Role = "user"
	}

	return user, nil
}
