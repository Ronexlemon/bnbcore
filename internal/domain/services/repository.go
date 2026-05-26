
package services

import (
    "context"
    "github.com/google/uuid"
)



type RoomServiceRepository interface {
    Create(ctx context.Context, rs *UnitService) (*UnitService, error)
    GetByID(ctx context.Context, id, tenantID uuid.UUID) (*UnitService, error)
    GetByUnit(ctx context.Context, unitID, tenantID uuid.UUID) ([]*UnitService, error)
    Update(ctx context.Context, id, tenantID uuid.UUID, req UpdateUnitServiceRequest) (*UnitService, error)
    Delete(ctx context.Context, id, tenantID uuid.UUID) error
}