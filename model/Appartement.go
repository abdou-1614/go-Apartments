package model

import (
	"gorm.io/gorm"
)

type Apartments struct {
	gorm.Model
	PropertyID string  `json:"propertyID"`
	Unit       string  `json:"unit"`
	Bedrooms   int     `json:"bedrooms"`
	Bathrooms  float32 `json:"bathrooms"`
}
