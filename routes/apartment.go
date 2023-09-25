package routes

import (
	"go-appointement/model"
	"go-appointement/storage"
	"go-appointement/utils"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
	"gorm.io/gorm/clause"
)

// GetApartmentByPropertyID retrieves apartments by property ID.
// @Summary Retrieve apartments by Property ID
// @Description Retrieve a list of apartments associated with a specific property.
// @Tags Apartments
// @Accept json
// @Produce json
// @Param id path int true "Property ID" Format(int64) Example(1)
// @Success 200 "List of apartments"
// @Failure 500 "Internal Server Error"
// @Router /apartments/property/{id} [get]
func GetApartmentByPropertyID(ctx iris.Context) {
	params := ctx.Params()
	id := params.Get("id")

	var apartment []model.Apartments

	apartmentExist := storage.DB.Where("property_id = ?", id).Find(&apartment)

	if apartmentExist.Error != nil {
		utils.CreateError(iris.StatusInternalServerError, "Error", apartmentExist.Error.Error(), ctx)
		return
	}

	ctx.JSON(apartment)
}

// UpdateApartment updates an apartment by ID.
// @Summary Update an apartment
// @Description Update an apartment by ID.
// @Tags Apartments
// @Accept json
// @Produce json
// @Param id path int true "Apartment ID" Format(int64) Example(1)
// @Param Authorization header string true "Bearer {token}" default(JWT Token)
// @Param input body []UpdateUnitsInput true "Apartment data to update"
// @Success 204 "No Content"
// @Failure 400 "Bad Request"
// @Failure 401 "Not Owner"
// @Failure 500 "Error"
// @Router /apartments/property/{id} [patch]
func UpdateApartment(ctx iris.Context) {
	params := ctx.Params()
	id := params.Get("id")

	property := GetPropertyAndAssociationsByPropertyID(id, ctx)

	if property == nil {
		return
	}

	claims := jwt.Get(ctx).(*utils.AccessToken)

	if property.UserID != claims.ID {
		utils.CreateError(iris.StatusBadRequest, "Not Owner", "Not Owner of Property", ctx)
		return
	}

	var updatedApartment []UpdateUnitsInput

	err := ctx.ReadJSON(&updatedApartment)

	if err != nil {
		utils.HandleValidationError(err, ctx)
		return
	}

	var newApartement []model.Apartments

	bedRoomsLow := property.BedroomLow
	bedRoomsHigh := property.BedroomHigh
	var bathroomHigh float32 = property.BathroomHigh
	var bathroomLow float32 = property.BathroomLow

	for _, apartment := range updatedApartment {
		if apartment.Bathrooms > bathroomHigh {
			bathroomHigh = apartment.Bathrooms
		}

		if apartment.Bathrooms < bathroomLow {
			bathroomLow = apartment.Bathrooms
		}

		if *apartment.Bedrooms > bedRoomsHigh {
			bedRoomsHigh = *apartment.Bedrooms
		}

		if *apartment.Bedrooms < bedRoomsLow {
			bedRoomsLow = *apartment.Bedrooms
		}

		currentApartment := model.Apartments{
			Unit:        apartment.Unit,
			Bedrooms:    *apartment.Bedrooms,
			Bathrooms:   apartment.Bathrooms,
			SqFt:        apartment.SqFt,
			Active:      apartment.Active,
			AvailableOn: apartment.AvailableOn,
			PropertyID:  property.ID,
		}

		if apartment.ID != nil {
			currentApartment.ID = *apartment.ID
			storage.DB.Model(&currentApartment).Updates(currentApartment)
		} else {
			newApartement = append(newApartement, currentApartment)
		}
	}

	if len(newApartement) > 0 {
		rowUpdated := storage.DB.Create(&newApartement)

		if rowUpdated.Error != nil {
			utils.CreateError(
				iris.StatusInternalServerError,
				"Error", rowUpdated.Error.Error(), ctx)
			return
		}
	}

	ctx.JSON(iris.StatusNoContent)
}

func GetPropertyAndAssociationsByPropertyID(id string, ctx iris.Context) *model.Property {

	var property model.Property
	propertyExists := storage.DB.Preload(clause.Associations).Find(&property, id)

	if propertyExists.Error != nil {
		utils.CreateInternalServerError(ctx)
		return nil
	}

	if propertyExists.RowsAffected == 0 {
		utils.CreateError(iris.StatusNotFound, "Not Found", "Not Found", ctx)
		return nil
	}

	return &property
}

type UpdateUnitsInput struct {
	ID          *uint     `json:"ID"`
	Unit        string    `json:"unit" validate:"max=512"`
	Bedrooms    *int      `json:"bedrooms" validate:"gte=0,max=6,required"`
	Bathrooms   float32   `json:"bathrooms" validate:"min=0.5,max=6.5,required"`
	SqFt        int       `json:"sqFt" validate:"max=100000000000,required"`
	Active      *bool     `json:"active" validate:"required"`
	AvailableOn time.Time `json:"availableOn" validate:"required"`
}
