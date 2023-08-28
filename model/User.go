package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	FirstName      string     `json:"firstName"`
	LastName       string     `json:"lastName"`
	Email          string     `json:"email"`
	Password       string     `json:"password"`
	SocialLogin    bool       `json:"socialLogin"`
	SocialProvider string     `json:"socialProvider"`
	Role           UserRole   `json:"role"`
	Properties     []Property `json:"properties"`
}

type UserRole string

const (
	RoleAdmin     UserRole = "admin"
	RoleUser      UserRole = "user"
	RoleLandlords UserRole = "landlords"
	RoleGuest     UserRole = "guest"
)
