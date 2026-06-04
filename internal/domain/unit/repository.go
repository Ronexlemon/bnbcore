package unit
import (
    "context"
    "github.com/google/uuid"
)



const (
    UnitStatusActive   UnitStatus = "active"
    UnitStatusInactive UnitStatus = "inactive"
    UnitStatusDeleted  UnitStatus = "deleted"
)



type UnitRepository interface {
    Create(ctx context.Context, unit *Unit) (*Unit, error)
    GetByID(ctx context.Context, id, tenantID uuid.UUID) (*Unit, error)
GetAll(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*Unit, error)
    Update(ctx context.Context, id, tenantID uuid.UUID, req UpdateUnitRequest) (*Unit, error)
    GetUnitByIdAndTenant(ctx context.Context, unitID uuid.UUID, tenantID uuid.UUID) (*Unit, error)
    GetImagesByUnitID(ctx context.Context, unitID uuid.UUID) ([]*UnitImage, error)
    Delete(ctx context.Context, id, tenantID uuid.UUID) error
    AddImage(ctx context.Context, image *UnitImage) (*UnitImage, error)
    RemoveImage(ctx context.Context, imageID, tenantID uuid.UUID) error
}