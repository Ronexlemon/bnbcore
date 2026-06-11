package tenant

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateTenant(ctx context.Context, shopName,ShopDescription, subdomain,phoneNumber string, userID uuid.UUID,long_Description string) (*Owner,error) {
	exists, err := s.repo.SubdomainExists(ctx, subdomain)
	if err != nil {
		return nil, fmt.Errorf("failed to check subdomain: %w", err)
	}
	if exists {
		return nil, errors.New("subdomain already taken")
	}
	return s.repo.CreateTenant(ctx, shopName,ShopDescription, subdomain,phoneNumber, userID,long_Description)
}

func (s *Service) GetTenantByID(ctx context.Context, id uuid.UUID) (*Tenant, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) GetTenantByUserID(ctx context.Context, userID uuid.UUID) (*Tenant, error) {
	return s.repo.FindByUserID(ctx, userID)
}

func (s *Service) GetTenantBySubdomain(ctx context.Context, subdomain string) (*Tenant, error) {
	return s.repo.FindBySubdomain(ctx, subdomain)
}

func (s *Service) UpdateTenant(ctx context.Context, id uuid.UUID, req UpdateTenantRequest) (*Tenant, error) {
	exists, err := s.repo.TenantExists(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to check tenant: %w", err)
	}
	if !exists {
		return nil, errors.New("tenant not found")
	}

	if req.Subdomain != nil {
		taken, err := s.repo.SubdomainExists(ctx, *req.Subdomain)
		if err != nil {
			return nil, fmt.Errorf("failed to check subdomain: %w", err)
		}
		if taken {
			return nil, errors.New("subdomain already taken")
		}
	}

	return s.repo.UpdateTenant(ctx, id, req)
}

func (s *Service) SubdomainExists(ctx context.Context, subdomain string) (bool, error) {
	subdomain = strings.ToLower(subdomain)
	return s.repo.SubdomainExists(ctx, subdomain)
}
func (s *Service) DeleteTenant(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteTenant(ctx, id)
}

func (s *Service) GetAll(ctx context.Context) ([]*Tenant, error) {
	return s.repo.GetAll(ctx)
}