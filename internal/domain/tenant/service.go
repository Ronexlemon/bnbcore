package tenant

import (
	"context"

	"github.com/ronexlemon/bnbcore/internal/auth/password"
)


type Service struct {
    repo Repository
    PasswordEngine *password.PasswordHasher
}

func NewService(r Repository,passEngine *password.PasswordHasher) *Service {
    return &Service{repo: r,PasswordEngine: passEngine}
}

func (s *Service) CreateTenant(ctx context.Context, name, subdomain string) error {
    return s.repo.CreateTenant(ctx, name, subdomain)
}

func(s *Service)  RegisterTenantWithUser(
	ctx context.Context,
	tenantName string,
	subdomain string,
	email string,
	password string,
) error {
    hashedPassword, err := s.PasswordEngine.Hash(password)
	if err != nil {
		return err
	}
    return s.repo.RegisterTenantWithUser(ctx,tenantName,subdomain,email,hashedPassword)
}