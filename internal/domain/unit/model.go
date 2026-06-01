package unit

import (
	"time"

	"github.com/google/uuid"
)

type UnitStatus string

type Unit struct {
    ID            uuid.UUID
    TenantID      uuid.UUID
    Title         string
    Description   string
    PricePerNight float64
    Name          string
    UnitType         string
    Location      string
    Latitude      float64
    Longitude     float64
    Status        UnitStatus
    Images        []UnitImage
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

type UnitImage struct {
    ID     uuid.UUID
    UnitID uuid.UUID
    URL    string
}

type CreateUnitRequest struct {
    Title         string   `json:"title"`
    Description   string   `json:"description"`
    Name          string   `json:"name"`
    UnitType          string   `json:"type"`
    PricePerNight float64  `json:"price_per_night"`
    Location      string   `json:"location"`
    Latitude      float64  `json:"latitude"`
    Longitude     float64  `json:"longitude"`
    Images        []string `json:"images"` // URLs
}

type UpdateUnitRequest struct {
    Title         *string    `json:"title"`
    Description   *string    `json:"description"`
    Name          *string    `json:"name"`
    UnitType         *string    `json:"type"`
    PricePerNight *float64   `json:"price_per_night"`
    Location      *string    `json:"location"`
    Latitude      *float64   `json:"latitude"`
    Longitude     *float64   `json:"longitude"`
    Status        *UnitStatus `json:"status"`
}