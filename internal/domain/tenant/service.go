package tenant

import (
	"context"
	"errors"
	"fmt"

	"github.com/ronexlemon/bnbcore/internal/auth/password"
	uuid "github.com/satori/go.uuid"
)


type Service struct {
    repo Repository
    PasswordEngine *password.PasswordHasher
}
var (
	ErrTenantNotFound      = errors.New("tenant not found")
	ErrSubdomainTaken      = errors.New("subdomain is already taken")
	ErrUserAlreadyOnboarded = errors.New("user already belongs to a tenant")
)

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

func (s *Service)SubdomainExists(ctx context.Context, subdomain string) (bool, error){
return s.repo.SubdomainExists(ctx,subdomain)
}
func (s *Service)CreateTenantWithOwner(ctx context.Context,userID uuid.UUID,tenantName string,subdomain string,) (*OnboardResult, error){
    taken, err := s.SubdomainExists(ctx, subdomain)
    if err != nil {
        return nil, fmt.Errorf("check subdomain: %w", err)
    }
    if taken {
        return nil, ErrSubdomainTaken
    }

    return s.repo.CreateTenantWithOwner(ctx,userID,tenantName,subdomain)

} 

func (s *Service)  FindBySubdomain(ctx context.Context, sub string)(*Tenant, error){
    return s.repo.FindBySubdomain(ctx,sub)
}