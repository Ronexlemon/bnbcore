package tenant

import (
	"context"

	"github.com/google/uuid"
)


type Repository interface {
    CreateTenant(ctx context.Context, name, subdomain string) error
     RegisterTenantWithUser(ctx context.Context,tenantName string,subdomain string,email string,password string) error
     CreateTenantWithOwner(ctx context.Context,userID uuid.UUID,tenantName string,subdomain string,) (*OnboardResult, error) 
     SubdomainExists(ctx context.Context, subdomain string) (bool, error)
     FindBySubdomain(ctx context.Context, sub string)(*Tenant, error)

	
}

