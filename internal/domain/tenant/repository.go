package tenant

import "context"


type Repository interface {
    CreateTenant(ctx context.Context, name, subdomain string) error
     RegisterTenantWithUser(ctx context.Context,tenantName string,subdomain string,email string,password string,
) error
	
}

