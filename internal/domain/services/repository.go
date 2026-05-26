
package services

import (
    "context"
    "github.com/google/uuid"
)



type RoomServiceRepository interface {
    Create(ctx context.Context, rs *RoomService) (*RoomService, error)
    GetByID(ctx context.Context, id, tenantID uuid.UUID) (*RoomService, error)
    GetByUnit(ctx context.Context, unitID, tenantID uuid.UUID) ([]*RoomService, error)
    Update(ctx context.Context, id, tenantID uuid.UUID, req UpdateRoomServiceRequest) (*RoomService, error)
    Delete(ctx context.Context, id, tenantID uuid.UUID) error
}