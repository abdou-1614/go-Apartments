package routes

import (
	"go-appointement/model"
	"go-appointement/storage"
	"go-appointement/utils"
	"math"
	"strconv"

	"github.com/kataras/iris/v12"
)

// CreateReview creates a review for a property.
// @Summary Create a review
// @Description Create a new review for a property by ID.
// @Tags Review
// @Accept json
// @Produce json
// @Param id path int true "Property ID" Format(int64)
// @Security JWT
// @Param input body CreateReviewInput true "Review data to create"
// @Success 200 "OK"
// @Failure 400 "Bad Request"
// @Failure 401 "Unauthorized"
// @Failure 500 "Internal Server Error"
// @Router /review/property{id} [post]
func CreateReview(ctx iris.Context) {
	params := ctx.Params()
	propertyID := params.Get("id")

	property := GetPropertyAndAssociationsByPropertyID(propertyID, ctx)

	if property == nil {
		return
	}

	var reviewInput CreateReviewInput

	err := ctx.ReadJSON(&reviewInput)

	if err != nil {
		utils.HandleValidationError(err, ctx)
		return
	}

	prpeID, cnvErr := strconv.ParseUint(propertyID, 10, 32)

	if cnvErr != nil {
		utils.CreateInternalServerError(ctx)
		return
	}

	review := model.Review{
		UserID:     reviewInput.UserID,
		PropertyID: uint(prpeID),
		Title:      reviewInput.Title,
		Body:       reviewInput.Body,
		Stars:      reviewInput.Stars,
	}

	storage.DB.Create(&review)

	updatePropertyReview(property, float32(review.Stars))

	ctx.JSON(review)
}

func updatePropertyReview(property *model.Property, stars float32) {
	var avg float32

	reviewLength := len(property.Reviews)

	if reviewLength == 0 {
		avg = stars
	} else {
		var sum float32

		for i := 0; i < len(property.Reviews); i++ {
			sum += float32(property.Reviews[i].Stars)
		}
		avg = (sum + stars) / (float32(reviewLength) + 1)
	}

	avg = float32(math.Round(float64(avg)*10) / 10)

	storage.DB.Model(&property).Update("stars", avg)
}

type CreateReviewInput struct {
	UserID uint   `json:"userID" validate:"required"`
	Body   string `json:"body" validate:"required"`
	Title  string `json:"title" validate:"required"`
	Stars  int    `json:"stars" validate:"required,gt=0,lt=6"`
}
