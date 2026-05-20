package tenant

import "context"


type Repository interface {
    CreateTenant(ctx context.Context, name, subdomain string) error
	
}

