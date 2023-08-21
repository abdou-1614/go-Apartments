package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Appartements struct {
	gorm.Model
	PropertyID  string         `json:"propertyID"`
	Unit        string         `json:"unit"`
	Bedrooms    int            `json:"bedrooms"`
	Bathrooms   float32        `json:"bathrooms"`
	SqFt        int            `json:"sqFt"`
	Rent        float32        `json:"rent"`
	Deposit     float32        `json:"deposit"`
	LeaseLength string         `json:"leaseLength"`
	AvailableOn time.Time      `json:"availableOn"`
	Active      *bool          `json:"active"`
	Description string         `json:"description"`
	Images      datatypes.JSON `json:"images"`
	Amenities   datatypes.JSON `json:"amenities"`
}
