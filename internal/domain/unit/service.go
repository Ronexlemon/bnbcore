package unit

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type UnitService struct {
    Repo UnitRepository
}

func NewUnitService(repo UnitRepository) *UnitService {
    return &UnitService{Repo: repo}
}

func (s *UnitService) CreateUnit(ctx context.Context, tenantID uuid.UUID, req CreateUnitRequest) (*Unit, error) {
    unit := &Unit{
        ID:            uuid.New(),
        TenantID:      tenantID,
        Title:         req.Title,
        Description:   req.Description,
        PricePerNight: req.PricePerNight,
        Name: req.Name,
        UnitType: req.UnitType,
        Location:      req.Location,
        Latitude:      req.Latitude,
        Longitude:     req.Longitude,
        Status:        UnitStatusActive,
    }
    return s.Repo.Create(ctx, unit)
}

func (s *UnitService) GetUnit(ctx context.Context, id, tenantID uuid.UUID) (*Unit, error) {
    return s.Repo.GetByID(ctx, id, tenantID)
}

func (s *UnitService) GetAllUnits(ctx context.Context, tenantID uuid.UUID) ([]*Unit, error) {
    return s.Repo.GetAll(ctx, tenantID)
}

func (s *UnitService) UpdateUnit(ctx context.Context, id, tenantID uuid.UUID, req UpdateUnitRequest) (*Unit, error) {
    return s.Repo.Update(ctx, id, tenantID, req)
}

func (s *UnitService) DeleteUnit(ctx context.Context, id, tenantID uuid.UUID) error {
    return s.Repo.Delete(ctx, id, tenantID)
}

func (s *UnitService) GetUnitImages(ctx context.Context, unitID uuid.UUID, tenantID uuid.UUID) ([]*UnitImage, error) {
    _, err := s.Repo.GetUnitByIdAndTenant(ctx, unitID, tenantID)
    if err != nil {
        return nil, fmt.Errorf("unit unauthorized or not found")
    }

    return s.Repo.GetImagesByUnitID(ctx, unitID)
}