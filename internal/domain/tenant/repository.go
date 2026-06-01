package tenant

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	
	CreateTenant(ctx context.Context, shopName,ShopDescription, subdomain string, userID uuid.UUID)(*Tenant,error)
	UpdateTenant(ctx context.Context, id uuid.UUID, req UpdateTenantRequest) (*Tenant, error)
	DeleteTenant(ctx context.Context, id uuid.UUID) error

	
	FindByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*Tenant, error)
	FindBySubdomain(ctx context.Context, sub string) (*Tenant, error)
	GetAll(ctx context.Context) ([]*Tenant, error)

	
	SubdomainExists(ctx context.Context, subdomain string) (bool, error)
	TenantExists(ctx context.Context, id uuid.UUID) (bool, error)
}

