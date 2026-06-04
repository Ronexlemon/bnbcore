package unit

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type UnitStatus string

type Unit struct {
    ID            uuid.UUID       `json:"id"`
    TenantID      uuid.UUID       `json:"tenant_id"`
    Title         string          `json:"title"`
    Description   string          `json:"description"`
    PricePerNight float64         `json:"price_per_night"`
    Name          string          `json:"name"`
    UnitType      string          `json:"type"`
    Location      string          `json:"location"`
    Latitude      float64         `json:"latitude"`
    Longitude     float64         `json:"longitude"`
    Adults        int32           `json:"adults"`
    Children      int32           `json:"children"`    
    Status        UnitStatus      `json:"status"`
    Amenities     json.RawMessage `json:"amenities"`
    Rules         json.RawMessage `json:"rules"` 
    Images        []UnitImage     `json:"images"`
    CreatedAt     time.Time       `json:"created_at"`
    UpdatedAt     time.Time       `json:"updated_at"`
}

type UnitImage struct {
    ID     uuid.UUID
    UnitID uuid.UUID
    URL    string
    ImageType string
}

type CreateUnitRequest struct {
    Title         string   `json:"title"`
    Description   string   `json:"description"`
    Name          string   `json:"name"`
    UnitType          string   `json:"type"`
    PricePerNight float64  `json:"price_per_night"`
    Location      string   `json:"location"`
    Latitude      float64  `json:"latitude"`
    Adults        int32    `json:"adults"`
    Children      int32   `json:"children"`
    Amenities     []string `json:"amenities"`
    Rules         []string `json:"rules"`
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
    Adults        *int32     `json:"adults"`
    Children      *int32     `json:"children"`
    NewAmenities  *[]string `json:"amenities"`
    NewRules      *[]string `json:"rules"`
   Amenities     json.RawMessage `json:"-"`
    Rules         json.RawMessage `json:"-"`
    Longitude     *float64   `json:"longitude"`
    Status        *UnitStatus `json:"status"`
}