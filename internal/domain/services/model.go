package services

import (
	"time"

	"github.com/google/uuid"
)

type UnitService struct {
    ID        uuid.UUID
    UnitID    uuid.UUID
    TenantID  uuid.UUID
    AgentName string
    Mobile    string
    Email     string
    IsActive  bool
    CreatedAt time.Time
    UpdatedAt time.Time
}

type CreateUnitServiceRequest struct {
    AgentName string `json:"agent_name"`
    Mobile    string `json:"mobile"`
    Email     string `json:"email"`
}

type UpdateUnitServiceRequest struct {
    AgentName *string `json:"agent_name"`
    Mobile    *string `json:"mobile"`
    Email     *string `json:"email"`
    IsActive  *bool   `json:"is_active"`
}