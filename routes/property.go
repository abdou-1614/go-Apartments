package routes

import (
	"encoding/json"
	"go-appointement/model"
	"go-appointement/storage"
	"go-appointement/utils"
	"strings"
	"time"

	"strconv"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
	"github.com/thanhpk/randstr"
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

func DeleteProperty(ctx iris.Context) {
	params := ctx.Params()
	id := params.Get("id")

	claims := jwt.Get(ctx).(*utils.AccessToken)

	var propertyModel model.Property

	if claims.ROLE == model.RoleAdmin {
		propertyDelete := storage.DB.Delete(&propertyModel, id)
		if propertyDelete.Error != nil {
			utils.CreateError(iris.StatusInternalServerError, "ERROR", propertyDelete.Error.Error(), ctx)
			return
		}
		ctx.JSON(iris.StatusNoContent)
		return
	}

	if claims.ID != propertyModel.UserID {
		utils.CreateError(iris.StatusForbidden, "NOT OWNER", "CAN'T DELETE PROPERTY", ctx)
		return
	}

	propertyDelete := storage.DB.Delete(propertyModel, id)

	if propertyDelete.Error != nil {
		utils.CreateError(iris.StatusInternalServerError, "Error", propertyDelete.Error.Error(), ctx)
		return
	}

	ctx.JSON(iris.StatusNoContent)
}

// UpdateProperty updates a property by ID.
// @Summary Update a property
// @Description Update a property by ID.
// @Tags Property
// @Accept json
// @Produce json
// @Param id path int true "Property ID" Format(int64)
// @Param Authorization header string true "Bearer {token}" default(JWT Token)
// @Param input body UpdatePropertyInput true "Property data to update"
// @Success 200 ""MESSAGE": "UPDATED SUCCCESS", "STATUS CODE": 200"
// @Failure 401 "CAN'T UPDATE PROPERTY"
// @Failure 500 "ERROR"
// @Router /update/{id} [put]
func UpdateProperty(ctx iris.Context) {
	params := ctx.Params()
	id := params.Get("id")

	var property model.Property

	claims := jwt.Get(ctx).(*utils.AccessToken)

	if claims.ID != property.UserID {
		utils.CreateError(iris.StatusForbidden, "NOT OWNER", "CAN'T UPDATE PROPERTY", ctx)
		return
	}

	propertyExist := storage.DB.Preload("Apartments").Find(&property, id)

	if propertyExist.Error != nil {
		utils.CreateError(iris.StatusInternalServerError, "ERROR", propertyExist.Error.Error(), ctx)
		return
	}

	var propertyInput UpdatePropertyInput

	err := ctx.ReadJSON(&propertyInput)

	if err != nil {
		utils.HandleValidationError(err, ctx)
		return
	}

	var newApartments []model.Apartments
	var newApartmentsImages []*[]string
	bedRoomsLow := property.BedroomLow
	bedRoomsHigh := property.BedroomHigh
	var bathRoomsLow float32 = property.BathroomLow
	var bathroomHigh float32 = property.BathroomHigh
	var rentLow float32 = propertyInput.Apartments[0].Rent
	var rentHigh float32 = propertyInput.Apartments[0].Rent

	for _, apartement := range propertyInput.Apartments {
		if apartement.Bathrooms < bathRoomsLow {
			bathRoomsLow = apartement.Bathrooms
		}

		if apartement.Bathrooms > bathroomHigh {
			bathroomHigh = apartement.Bathrooms
		}

		if *apartement.Bedrooms < bedRoomsLow {
			bedRoomsLow = *apartement.Bedrooms
		}

		if *apartement.Bedrooms > bedRoomsHigh {
			bedRoomsHigh = *apartement.Bedrooms
		}

		if apartement.Rent < rentLow {
			rentLow = apartement.Rent
		}

		if apartement.Rent > rentHigh {
			rentHigh = apartement.Rent
		}

		amenities, _ := json.Marshal(apartement.Amenities)

		currApartement := model.Apartments{
			Unit:        apartement.Unit,
			Bedrooms:    *apartement.Bedrooms,
			Bathrooms:   apartement.Bathrooms,
			PropertyID:  property.ID,
			SqFt:        apartement.SqFt,
			Rent:        apartement.Rent,
			Deposit:     *apartement.Deposit,
			LeaseLength: apartement.LeaseLength,
			AvailableOn: apartement.AvailableOn,
			Active:      apartement.Active,
			Amenities:   amenities,
			Description: apartement.Description,
		}
		if apartement.ID != nil {
			currApartement.ID = *apartement.ID
			UpdateApartmentsAndImage(currApartement, apartement.Images)
		} else {
			newApartments = append(newApartments, currApartement)
			newApartmentsImages = append(newApartmentsImages, &apartement.Images)
		}
	}
	storage.DB.Create(&newApartments)

	for index, apartement := range newApartments {
		if len(*newApartmentsImages[index]) > 0 {
			UpdateApartmentsAndImage(apartement, *newApartmentsImages[index])
		}
	}
	propertyAmenities, _ := json.Marshal(propertyInput.Amenities)
	includedUtilities, _ := json.Marshal(propertyInput.IncludedUtilities)

	property.UnitType = propertyInput.UnitType
	property.Description = propertyInput.Description
	property.IncludedUtilities = includedUtilities
	property.PetsAllowed = propertyInput.PetsAllowed
	property.LaundryType = propertyInput.LaundryType
	property.ParkingFee = *propertyInput.ParkingFee
	property.Amenities = propertyAmenities
	property.Name = propertyInput.Name
	property.FirstName = propertyInput.FirstName
	property.LastName = propertyInput.LastName
	property.Email = propertyInput.Email
	property.CallingCode = propertyInput.CallingCode
	property.CountryCode = propertyInput.CountryCode
	property.PhoneNumber = propertyInput.PhoneNumber
	property.Website = propertyInput.Website
	property.OnMarket = propertyInput.OnMarket
	property.BathroomHigh = bathroomHigh
	property.BathroomLow = bathRoomsLow
	property.BedroomLow = bedRoomsLow
	property.BedroomLow = bedRoomsHigh
	property.RentLow = rentLow
	property.RentHigh = rentHigh

	imagesArr := insertImages(InsertImages{
		images:     propertyInput.Images,
		propertyID: strconv.FormatUint(uint64(property.ID), 10),
	})

	imageJson, _ := json.Marshal(imagesArr)

	property.Images = imageJson

	rowsUpdated := storage.DB.Model(&property).Updates(property)

	if rowsUpdated.Error != nil {
		utils.CreateError(iris.StatusInternalServerError, "ERROR", rowsUpdated.Error.Error(), ctx)
		return
	}

	response := map[string]interface{}{
		"MESSAGE":     "UPDATED SUCCCESS",
		"STATUS CODE": iris.StatusOK,
	}

	ctx.JSON(response)
}

// GetAllProperty
// @Summary Get All properties
// @Description Retrieves All properties.
// @Accept json
// @Produce json
// @Success 200 {array} PropertyResponse
// @Tags Property
// @Failure 500 "Internal Server Error"
// @Router /getAllProperties [get]
func GetAllProperty(ctx iris.Context) {
	var query PaginationQuery

	if err := ctx.ReadQuery(&query); err != nil {
		utils.HandleValidationError(err, ctx)
	}

	if query.Page == 0 {
		query.Page = 1
	}

	if query.PerPage == 0 {
		query.PerPage = 20
	}

	offSet := (query.Page - 1) * query.PerPage

	var property []model.Property

	if err := storage.DB.Offset(offSet).Limit(query.PerPage).Find(&property).Error; err != nil {
		utils.CreateInternalServerError(ctx)
		return
	}

	ctx.JSON(property)
}

// GetTopRatedProperty
// @Summary Get top-rated properties
// @Description Retrieves the top-rated properties in descending order.
// @Accept json
// @Produce json
// @Success 200 {array} PropertyResponse
// @Tags Property
// @Failure 500 "Internal Server Error"
// @Router /getTop [get]
func GetTopRatedPropert(ctx iris.Context) {
	var property []model.Property
	if err := storage.DB.
		Select("properties.*, AVG(reviews.stars) AS average_stars").
		Joins("LEFT JOIN reviews ON properties.id = reviews.property_id").
		Group("properties.id").
		Order("average_stars DESC").
		Limit(10).
		Find(&property).
		Error; err != nil {
		utils.CreateInternalServerError(ctx)
		return
	}
	ctx.JSON(property)
}

func UpdateApartmentsAndImage(apartment model.Apartments, images []string) {
	apartmentID := strconv.FormatUint(uint64(apartment.ID), 10)

	apartmentImage := insertImages(InsertImages{
		images:      images,
		propertyID:  strconv.FormatUint(uint64(apartment.PropertyID), 10),
		apartmentID: &apartmentID,
	})

	if len(apartmentImage) > 0 {
		imageJson, _ := json.Marshal(apartmentImage)
		apartment.Images = imageJson
	}

	storage.DB.Model(&apartment).Updates(apartment)
}

func insertImages(arg InsertImages) []string {
	var imagesArr []string

	for _, image := range arg.images {
		if !strings.Contains(image, storage.BucketName) {
			imageID := randstr.Hex(16)
			imageStr := "property/" + arg.propertyID
			if arg.apartmentID != nil {
				imageStr += "/apartment/" + *arg.apartmentID
			}
			imageStr += "/" + imageID

			urlMap := storage.UploadBaseImage(image, imageStr)
			imagesArr = append(imagesArr, urlMap["url"])
		} else {
			imagesArr = append(imagesArr, image)
		}
	}

	return imagesArr
}

type InsertImages struct {
	images      []string
	propertyID  string
	apartmentID *string
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

type UpdatePropertyInput struct {
	UnitType          string                  `json:"unitType" validate:"required,oneof=single multiple"`
	Description       string                  `json:"description"`
	Images            []string                `json:"images"`
	IncludedUtilities []string                `json:"includedUtilities"`
	PetsAllowed       string                  `json:"petsAllowed" validate:"required"`
	LaundryType       string                  `json:"laundryType" validate:"required"`
	ParkingFee        *float32                `json:"parkingFee"`
	Amenities         []string                `json:"amenities"`
	Name              string                  `json:"name"`
	FirstName         string                  `json:"firstName"`
	LastName          string                  `json:"lastName"`
	Email             string                  `json:"email" validate:"required,email"`
	CallingCode       string                  `json:"callingCode"`
	CountryCode       string                  `json:"countryCode"`
	PhoneNumber       string                  `json:"phoneNumber" validate:"required"`
	Website           string                  `json:"website" validate:"omitempty,url"`
	OnMarket          *bool                   `json:"onMarket" validate:"required"`
	Apartments        []UpdateApartmentsInput `json:"apartments" validate:"required,dive"`
}

type PaginationQuery struct {
	Page    int `json:"page" validate:"gte=1"`
	PerPage int `json:"perPage" validate:"gte=1"`
}

type UpdateApartmentsInput struct {
	ID          *uint     `json:"ID"`
	Unit        string    `json:"unit" validate:"required,max=256"`
	Bedrooms    *int      `json:"bedroom" validate:"required,gte=0,max=6"`
	Bathrooms   float32   `json:"bathrooms" validate:"min=0.5,max=6.5,required"`
	SqFt        int       `json:"sqFt" validate:"max=100000000000,required"`
	Rent        float32   `json:"rent" validate:"required"`
	Deposit     *float32  `json:"deposit" validate:"required"`
	LeaseLength string    `json:"leaseLength" validate:"required,max=256"`
	AvailableOn time.Time `json:"availableOn" validate:"required"`
	Active      *bool     `json:"active" validate:"required"`
	Images      []string  `json:"images"`
	Amenities   []string  `json:"amenities"`
	Description string    `json:"description"`
}

type ApartmentsInput struct {
	Unit      string  `json:"unit" validate:"required,max=256"`
	Bedrooms  *int    `json:"bedroom" validate:"required,gte=0,max=6"`
	Bathrooms float32 `json:"bathrooms" validate:"min=0.5,max=6.5,required"`
}

type PropertyResponse struct {
	ID           uint    `json:"id"`
	UnitType     string  `json:"unitType"`
	PropertyType string  `json:"propertyType"`
	Street       string  `json:"street"`
	City         string  `json:"city"`
	State        string  `json:"state"`
	Zip          int     `json:"zip"`
	Lat          float32 `json:"lat"`
	Lng          float32 `json:"lng"`
	Stars        int     `json:"stars"`
}
