package routes

import (
	"go-appointement/model"
	"go-appointement/storage"
	"go-appointement/utils"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
)

func SubmitRoleChangeRequest(ctx iris.Context) {
	var requestStatus RoleChangeRequest

	err := ctx.ReadJSON(&requestStatus)

	if err != nil {
		utils.HandleValidationError(err, ctx)
		return
	}

	claims := jwt.Get(ctx).(*utils.AccessToken)

	if claims.ID != requestStatus.UserID {
		utils.CreateError(iris.StatusInternalServerError, "Not Correct ID", "You can only request a role change for yourself", ctx)
		return
	}
	var roleChangeRequest model.RoleChangeRequest
	roleChangeRequest.UserID = requestStatus.UserID
	roleChangeRequest.Status = model.RequestPending

	if err := storage.DB.Create(&roleChangeRequest).Error; err != nil {
		utils.CreateInternalServerError(ctx)
		return
	}

	ctx.JSON(iris.StatusCreated)
}

type RoleChangeRequest struct {
	UserID uint `json:"userID" validate:"required"`
}
