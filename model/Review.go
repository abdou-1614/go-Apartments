package model

import "gorm.io/gorm"

type Review struct {
	gorm.Model
	UserID     uint   `json:"userID"`
	PropertyID uint   `json:"propertyID"`
	Body       string `json:"body"`
	Title      string `json:"title"`
	Stars      int    `json:"stars"`
}
