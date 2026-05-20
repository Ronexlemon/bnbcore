package tenant

import "context"


type Service struct {
    repo Repository
}

func NewService(r Repository) *Service {
    return &Service{repo: r}
}

func (s *Service) CreateTenant(ctx context.Context, name, subdomain string) error {
    return s.repo.CreateTenant(ctx, name, subdomain)
}