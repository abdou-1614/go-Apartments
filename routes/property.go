package routes

import "github.com/kataras/iris/v12"

func CreatePropert(ctx iris.Context) {}

type PropertyInput struct {
	UnitType     string             `json:"unitType" validate:"required, oneof= single multiple"`
	PropertType  string             `json:"propertyType" validate:"required, max=256"`
	Street       string             `json:"street" validate:"required, max=512"`
	City         string             `json:"city" validate:"required, max=256"`
	State        string             `json:"state" validate:"required, max=256"`
	Zip          int                `json:"zip" validate:"required"`
	Lat          float32            `json:"lat" validate:"required"`
	Lng          float32            `json:"lng" validate:"required"`
	ManagerID    uint               `json:"managerID" validate:"required"`
	Appartements []AppartementInput `json:"appartements" validate:"required, dive"`
}

type AppartementInput struct {
	Unit      string  `json:"unit" validate:"required max=256"`
	Badrooms  *int    `json:"badroom" validate:"required, gte=0, max=6"`
	Bathrooms float32 `json:"bathrooms" validate:"min=0.5, max=6.5, required"`
}
