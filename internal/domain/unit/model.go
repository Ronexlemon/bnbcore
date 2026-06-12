package unit

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/ronexlemon/bnbcore/internal/domain/tenant"
)

type UnitStatus string
type UnitType string

const (
	UnitStatusActive      UnitStatus = "active"
	UnitStatusInactive    UnitStatus = "inactive"
	UnitStatusDeleted     UnitStatus = "deleted"
	UnitStatusMaintenance UnitStatus = "maintenance"
)

type Unit struct {
	ID               uuid.UUID       `json:"id"`
	TenantID         uuid.UUID       `json:"tenant_id"`

	// Identity
	Title            string          `json:"title"`
	Name             string          `json:"name"`
	ShortDescription string          `json:"short_description"`
	Description      string          `json:"description"`
	Slug             string          `json:"slug"`
	UnitType         string          `json:"type"`

	// Capacity
	Guests           int32           `json:"guests"`
	Bedrooms         int32           `json:"bedrooms"`
	Beds             int32           `json:"beds"`
	Bathrooms        int32           `json:"bathrooms"`

	// Location
	Location         string          `json:"location"`
	Latitude         float64         `json:"latitude"`
	Longitude        float64         `json:"longitude"`
	ApartmentName    string          `json:"apartment_name"`
	HouseNumber      string          `json:"house_number"`
	Floor            string          `json:"floor"`
	AccessNote       string          `json:"access_note"`

	// Pricing
	PriceWeekday     float64         `json:"price_weekday"`
	PriceWeekend     float64         `json:"price_weekend"`

	// Check-in / Check-out
	CheckinTime      string          `json:"checkin_time"`  
	CheckoutTime     string          `json:"checkout_time"` 

	// Flexible content
	Amenities        json.RawMessage `json:"amenities"`
	Rules            json.RawMessage `json:"rules"`

	// Contact & status
	PhoneNumber      string          `json:"phone_number"`
	Status           UnitStatus      `json:"status"`

	Images           []UnitImage     `json:"images"`

	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type UnitImage struct {
	ID        uuid.UUID `json:"id"`
	UnitID    uuid.UUID `json:"unit_id"`
	URL       string    `json:"url"`
	ImageType string    `json:"image_type"`
	SortOrder int       `json:"sort_order"`
}

type CreateUnitRequest struct {
	// Identity
	Title            string   `json:"title"`
	Name             string   `json:"name"`
	ShortDescription string   `json:"short_description"`
	Description      string   `json:"description"`
	UnitType         string   `json:"type"`

	// Capacity
	Guests           int32    `json:"guests"`
	Bedrooms         int32    `json:"bedrooms"`
	Beds             int32    `json:"beds"`
	Bathrooms        int32    `json:"bathrooms"`

	// Location
	Location         string   `json:"location"`
	Latitude         float64  `json:"latitude"`
	Longitude        float64  `json:"longitude"`
	ApartmentName    string   `json:"apartment_name"`
	HouseNumber      string   `json:"house_number"`
	Floor            string   `json:"floor"`
	AccessNote       string   `json:"access_note"`

	// Pricing
	PriceWeekday     float64  `json:"price_weekday"`
	PriceWeekend     float64  `json:"price_weekend"`

	// Check-in / Check-out
	CheckinTime      string   `json:"checkin_time"`
	CheckoutTime     string   `json:"checkout_time"`

	// Flexible content
	Amenities        []string `json:"amenities"`
	Rules            []string `json:"rules"`

	// Contact
	PhoneNumber      string   `json:"phone_number"`

	Images           []string `json:"images"` // URLs
}

type UpdateUnitRequest struct {
	// Identity
	Title            *string     `json:"title"`
	Name             *string     `json:"name"`
	ShortDescription *string     `json:"short_description"`
	Description      *string     `json:"description"`
	Slug             *string     `json:"slug"`
	UnitType         *string     `json:"type"`

	// Capacity
	Guests           *int32      `json:"guests"`
	Bedrooms         *int32      `json:"bedrooms"`
	Beds             *int32      `json:"beds"`
	Bathrooms        *int32      `json:"bathrooms"`

	// Location
	Location         *string     `json:"location"`
	Latitude         *float64    `json:"latitude"`
	Longitude        *float64    `json:"longitude"`
	ApartmentName    *string     `json:"apartment_name"`
	HouseNumber      *string     `json:"house_number"`
	Floor            *string     `json:"floor"`
	AccessNote       *string     `json:"access_note"`

	// Pricing
	PriceWeekday     *float64    `json:"price_weekday"`
	PriceWeekend     *float64    `json:"price_weekend"`

	// Check-in / Check-out
	CheckinTime      *string     `json:"checkin_time"`
	CheckoutTime     *string     `json:"checkout_time"`

	// Flexible content — incoming as string slices, serialized before DB write
	NewAmenities     *[]string   `json:"amenities"`
	NewRules         *[]string   `json:"rules"`
	Amenities        json.RawMessage `json:"-"`
	Rules            json.RawMessage `json:"-"`

	// Contact & status
	PhoneNumber      *string     `json:"phone_number"`
	Status           *UnitStatus `json:"status"`
}

type HostUnitsResult struct {
	Tenant *tenant.Tenant `json:"tenant"`
	Units  []*Unit   `json:"units"`
}