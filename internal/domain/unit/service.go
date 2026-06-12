package unit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/ronexlemon/bnbcore/internal/domain/tenant"
)

type UnitService struct {
	Repo UnitRepository
	TenantService *tenant.Service
}

func NewUnitService(repo UnitRepository,tenant_service *tenant.Service) *UnitService {
	return &UnitService{Repo: repo,TenantService: tenant_service}
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
		return nil, fmt.Errorf("failed to marshal amenities: %w", err)
	}
	rulesJSON, err := json.Marshal(rules)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal rules: %w", err)
	}

	slug, err := s.Repo.GenerateUniqueSlug(ctx, tenantID, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to generate slug: %w", err)
	}

	unit := &Unit{
		ID:       uuid.New(),
		TenantID: tenantID,

		// Identity
		Title:            req.Title,
		Name:             req.Name,
		ShortDescription: req.ShortDescription,
		Description:      req.Description,
		Slug:             slug,
		UnitType:         req.UnitType,

		// Capacity
		Guests:    req.Guests,
		Bedrooms:  req.Bedrooms,
		Beds:      req.Beds,
		Bathrooms: req.Bathrooms,

		// Location
		Location:      req.Location,
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		ApartmentName: req.ApartmentName,
		HouseNumber:   req.HouseNumber,
		Floor:         req.Floor,
		AccessNote:    req.AccessNote,

		// Pricing
		PriceWeekday: req.PriceWeekday,
		PriceWeekend: req.PriceWeekend,

		// Check-in / Check-out
		CheckinTime:  req.CheckinTime,
		CheckoutTime: req.CheckoutTime,

		// Flexible content
		Amenities: json.RawMessage(amenitiesJSON),
		Rules:     json.RawMessage(rulesJSON),

		// Contact & status
		PhoneNumber: req.PhoneNumber,
		Status:      UnitStatusActive,
	}

	return s.Repo.Create(ctx, unit)
}

func (s *UnitService) GetUnit(ctx context.Context, id, tenantID uuid.UUID) (*Unit, error) {
	return s.Repo.GetByID(ctx, id, tenantID)
}

func (s *UnitService) GetBySlug(ctx context.Context, slug string, tenantID uuid.UUID) (*Unit, error) {
	return s.Repo.GetBySlug(ctx, slug, tenantID)
}

func (s *UnitService) GetUnitBySlugAndTenant(ctx context.Context, slug string, tenantID uuid.UUID) (*Unit, error) {
	return s.Repo.GetUnitBySlugAndTenant(ctx, slug, tenantID)
}

func (s *UnitService) GetAllUnits(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*Unit, error) {
	return s.Repo.GetAll(ctx, tenantID, limit, offset)
}
func (s *UnitService) GetHostUnitsDetails(ctx context.Context, tenantID uuid.UUID, limit, offset int) (*HostUnitsResult, error) {
	t, err := s.TenantService.GetTenantByID(ctx, tenantID) 
	if err != nil {
		return nil, err
	}

	units, err := s.Repo.GetAll(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}

	return &HostUnitsResult{Tenant: t, Units: units}, nil
}

func (s *UnitService) UpdateUnit(ctx context.Context, id, tenantID uuid.UUID, req UpdateUnitRequest) (*Unit, error) {
	if req.NewAmenities != nil {
		items := *req.NewAmenities
		if items == nil {
			items = []string{}
		}
		b, err := json.Marshal(items)
		if err != nil {
			return nil, fmt.Errorf("failed to process amenities: %w", err)
		}
		req.Amenities = json.RawMessage(b)
	}

	if req.NewRules != nil {
		items := *req.NewRules
		if items == nil {
			items = []string{}
		}
		b, err := json.Marshal(items)
		if err != nil {
			return nil, fmt.Errorf("failed to process rules: %w", err)
		}
		req.Rules = json.RawMessage(b)
	}

	// Regenerate slug only if name is being changed
	if req.Name != nil {
		slug, err := s.Repo.GenerateUniqueSlug(ctx, tenantID, *req.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to generate slug: %w", err)
		}
		req.Slug = &slug
	}

	return s.Repo.Update(ctx, id, tenantID, req)
}

func (s *UnitService) DeleteUnit(ctx context.Context, id, tenantID uuid.UUID) error {
	return s.Repo.Delete(ctx, id, tenantID)
}

func (s *UnitService) GetUnitImages(ctx context.Context, unitID, tenantID uuid.UUID) ([]*UnitImage, error) {
	_, err := s.Repo.GetUnitByIdAndTenant(ctx, unitID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("unit unauthorized or not found")
	}
	return s.Repo.GetImagesByUnitID(ctx, unitID)
}