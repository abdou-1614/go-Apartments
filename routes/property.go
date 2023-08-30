package routes

import (
	"go-appointement/model"
	"go-appointement/storage"
	"go-appointement/utils"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
)

func CreateProperty(ctx iris.Context) {
	var propertyInput PropertyInput
	err := ctx.ReadJSON(&propertyInput)

	if err != nil {
		utils.HandleValidationError(err, ctx)
		return
	}

	claims := jwt.Get(ctx).(*utils.AccessToken)

	if claims.ID != propertyInput.UserID {
		utils.CreateError(iris.StatusBadRequest, "Not Owner", "Not Owner of Property", ctx)
		return
	}
	var appartements []model.Apartments
	bedroomsLow := 0
	bedroomsHigh := 0
	var bathroomsLow float32 = 0.5
	var bathroomsHigh float32 = 0.5

	for _, element := range propertyInput.Apartments {
		if element.Bathrooms < bathroomsLow {
			bathroomsLow = element.Bathrooms
		}

		if element.Bathrooms > bathroomsHigh {
			bathroomsHigh = element.Bathrooms
		}

		if *element.Bedrooms < bedroomsLow {
			bedroomsLow = *element.Bedrooms
		}

		if *element.Bedrooms > bedroomsHigh {
			bedroomsHigh = *element.Bedrooms
		}

		appartements = append(appartements, model.Apartments{
			Unit:      element.Unit,
			Bedrooms:  *element.Bedrooms,
			Bathrooms: element.Bathrooms,
		})
	}
	property := model.Property{
		UnitType:     propertyInput.UnitType,
		PropertyType: propertyInput.PropertType,
		Street:       propertyInput.Street,
		City:         propertyInput.City,
		State:        propertyInput.State,
		Zip:          propertyInput.Zip,
		Lat:          propertyInput.Lat,
		Lng:          propertyInput.Lng,
		BedroomLow:   bedroomsLow,
		BedroomHigh:  bedroomsHigh,
		BathroomLow:  bathroomsLow,
		BathroomHigh: bathroomsHigh,
		Apartments:   appartements,
		UserID:       propertyInput.UserID,
	}

	storage.DB.Create(&property)
	ctx.JSON(property)
}

func GetProperty(ctx iris.Context) {
	params := ctx.Params()
	id := params.Get("id")

	var property model.Property

	propertyExist := storage.DB.Preload("Apartments").Find(&property, id)

	if propertyExist.Error != nil {
		utils.CreateError(iris.StatusInternalServerError, "Error", propertyExist.Error.Error(), ctx)
		return
	}

	if propertyExist.RowsAffected == 0 {
		utils.CreateError(iris.StatusNotFound, "Property Not Exist", "Property Not Exist", ctx)
		return
	}

	ctx.JSON(property)

}

type PropertyInput struct {
	UnitType    string            `json:"unitType" validate:"required,oneof=single multiple"`
	PropertType string            `json:"propertyType" validate:"required,max=256"`
	Street      string            `json:"street" validate:"required,max=512"`
	City        string            `json:"city" validate:"required,max=256"`
	State       string            `json:"state" validate:"required,max=256"`
	Zip         int               `json:"zip" validate:"required"`
	Lat         float32           `json:"lat" validate:"required"`
	Lng         float32           `json:"lng" validate:"required"`
	UserID      uint              `json:"userID" validate:"required"`
	Apartments  []ApartmentsInput `json:"apartments" validate:"required,dive"`
}

type ApartmentsInput struct {
	Unit      string  `json:"unit" validate:"required,max=256"`
	Bedrooms  *int    `json:"bedroom" validate:"required,gte=0,max=6"`
	Bathrooms float32 `json:"bathrooms" validate:"min=0.5,max=6.5,required"`
}
