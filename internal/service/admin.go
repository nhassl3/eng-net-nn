package service

import (
	"context"
	"fmt"

	"github.com/nhassl3/IpBuild-backend/internal/repository/postgres"
)

type AdminService struct {
	adminRepo postgres.Admin
}

func NewAdminService(adminRepo postgres.Admin) *AdminService {
	return &AdminService{
		adminRepo: adminRepo,
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
