package unit

import (
	"context"
	"encoding/json"
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
amenities := req.Amenities
    if amenities == nil {
        amenities = []string{}
    }

    rules := req.Rules
    if rules == nil {
        rules = []string{}
    }
    
    amenitiesJSON, err := json.Marshal(amenities)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal unit amenities to jsonb: %w", err)
    }

    rulesJSON, err := json.Marshal(rules)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal unit rules to jsonb: %w", err)
    }
    
    unit := &Unit{
        ID:            uuid.New(),
        TenantID:      tenantID,
        Title:         req.Title,
        Description:   req.Description,
        PricePerNight: req.PricePerNight,
        Name:          req.Name,
        UnitType:      req.UnitType,
        Location:      req.Location,
        Latitude:      req.Latitude,
        Longitude:     req.Longitude,
        Status:        UnitStatusActive,
        Adults: req.Adults,
        Children: req.Children,
        Amenities:     json.RawMessage(amenitiesJSON), 
        Rules:         json.RawMessage(rulesJSON),    
    }
    return s.Repo.Create(ctx, unit)
}

func (s *UnitService) GetUnit(ctx context.Context, id, tenantID uuid.UUID) (*Unit, error) {
    return s.Repo.GetByID(ctx, id, tenantID)
}

func (s *UnitService) GetAllUnits(ctx context.Context, tenantID uuid.UUID,limit,offset int) ([]*Unit, error) {
    return s.Repo.GetAll(ctx, tenantID,limit,offset)
}

func (s *UnitService) UpdateUnit(ctx context.Context, id, tenantID uuid.UUID, req UpdateUnitRequest) (*Unit, error) {
    if req.NewAmenities != nil {
        items := *req.NewAmenities
        if items == nil {
            items = []string{}
        }
        bytes, err := json.Marshal(items)
        if err != nil {
            return nil, fmt.Errorf("failed to process amenities updates: %w", err)
        }
        req.Amenities = json.RawMessage(bytes)
    }
    if req.NewRules != nil {
        items := *req.NewRules
        if items == nil {
            items = []string{}
        }
        bytes, err := json.Marshal(items)
        if err != nil {
            return nil, fmt.Errorf("failed to process rules updates: %w", err)
        }
        req.Rules = json.RawMessage(bytes)
    }
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