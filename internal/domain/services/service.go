package services

import (
    "context"
    "errors"

    "github.com/google/uuid"
)

type Service struct {
    Repo RoomServiceRepository
}

func NewService(repo RoomServiceRepository) *Service {
    return &Service{Repo: repo}
}

func (s *Service) Create(ctx context.Context, unitID, tenantID uuid.UUID, req CreateUnitServiceRequest) (*UnitService, error) {
    if req.AgentName == "" || req.Mobile == "" {
        return nil, errors.New("agent_name and mobile are required")
    }

    rs := &UnitService{
        ID:        uuid.New(),
        UnitID:    unitID,
        TenantID:  tenantID,
        AgentName: req.AgentName,
        Mobile:    req.Mobile,
        Email:     req.Email,
        IsActive:  true,
    }
    return s.Repo.Create(ctx, rs)
}

func (s *Service) GetByUnit(ctx context.Context, unitID, tenantID uuid.UUID) ([]*UnitService, error) {
    return s.Repo.GetByUnit(ctx, unitID, tenantID)
}

func (s *Service) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*UnitService, error) {
    return s.Repo.GetByID(ctx, id, tenantID)
}

func (s *Service) Update(ctx context.Context, id, tenantID uuid.UUID, req UpdateUnitServiceRequest) (*UnitService, error) {
    return s.Repo.Update(ctx, id, tenantID, req)
}

func (s *Service) Delete(ctx context.Context, id, tenantID uuid.UUID) error {
    return s.Repo.Delete(ctx, id, tenantID)
}