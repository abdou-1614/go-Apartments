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

	response := map[string]interface{}{
		"MESSAGE": "THE REQUEST HAS BEEN SUBMITED SUCCESSFULLY",
		"STATUS":  iris.StatusCreated,
	}

	ctx.JSON(response)
}

func ManageRoleChangeRequests(ctx iris.Context) {
	claims := jwt.Get(ctx).(*utils.AccessToken)

	if claims.ROLE != model.RoleAdmin {
		utils.CreateError(iris.StatusForbidden, "NOT ADMIN", "Only admins can manage role change requests", ctx)
		return
	}

	var requests []model.RoleChangeRequest

	if err := storage.DB.Where("status = ?", model.RequestPending).Preload("User").Find(&requests).Error; err != nil {
		utils.CreateInternalServerError(ctx)
		return
	}

	var response []RoleChangeRequestWithUser

	for _, request := range requests {
		response = append(response, RoleChangeRequestWithUser{
			ID:       request.ID,
			UserID:   request.UserID,
			UserName: request.User.FirstName + " " + request.User.LastName,
			NewRole:  request.User.Role,
			Status:   request.Status,
		})
	}

	ctx.JSON(response)
}

func AcceptRoleChangeRequest(ctx iris.Context) {
	claims := jwt.Get(ctx).(*utils.AccessToken)

	if claims.ROLE != model.RoleAdmin {
		utils.CreateError(iris.StatusForbidden, "NOT ADMIN", "Only admins can manage role change requests", ctx)
		return
	}

	requestID, err := ctx.Params().GetUint("id")

	if err != nil {
		utils.CreateError(iris.StatusBadRequest, "INVALID ID", "INVLAID ID", ctx)
		return
	}

	var request model.RoleChangeRequest

	if err := storage.DB.First(&request, requestID).Error; err != nil {
		utils.CreateError(iris.StatusInternalServerError, "Faild to find", "Faild to Find request", ctx)
		return
	}

	request.Status = model.RequestAccepted

	if err := storage.DB.Save(&request).Error; err != nil {
		utils.CreateError(iris.StatusInternalServerError, "Faild To Save", "Faild to save in DB", ctx)
		return
	}

	var user model.User

	if err := storage.DB.First(&user, request.UserID); err != nil {
		utils.CreateError(iris.StatusInternalServerError, "Faild To Find", "Faild to fetch user", ctx)
		return
	}

	user.Role = model.RoleAdmin

	if err := storage.DB.Save(&user).Error; err != nil {
		utils.CreateError(iris.StatusInternalServerError, "Faild To save new role of user", "Faild To save new role of user in DB", ctx)
		return
	}

	response := map[string]interface{}{
		"MESSAGE": "THE REQUEST HAS BEEN ACCEPTED SUCCESSFULLY",
		"STATUS":  iris.StatusOK,
	}

	ctx.JSON(response)
}

func RejectRoleChangeRequest(ctx iris.Context) {
	claims := jwt.Get(ctx).(*utils.AccessToken)

	if claims.ROLE != model.RoleAdmin {
		utils.CreateError(iris.StatusForbidden, "NOT ADMIN", "Only admins can manage role change requests", ctx)
		return
	}

	requestID, err := ctx.Params().GetUint("id")

	if err != nil {
		utils.CreateError(iris.StatusBadRequest, "INVALID ID", "INVLAID ID", ctx)
		return
	}

	var request model.RoleChangeRequest

	if err := storage.DB.First(&request, requestID).Error; err != nil {
		utils.CreateError(iris.StatusInternalServerError, "Faild to find", "Faild to Find request", ctx)
		return
	}

	request.Status = model.RequestRejected

	if err := storage.DB.Save(&request).Error; err != nil {
		utils.CreateError(iris.StatusInternalServerError, "Faild To Save", "Faild to save in DB", ctx)
		return
	}

	response := map[string]interface{}{
		"MESSAGE": "THE REQUEST HAS BEEN REJECTED SUCCESSFULLY",
		"STATUS":  iris.StatusOK,
	}

	ctx.JSON(response)
}

type RoleChangeRequest struct {
	UserID uint `json:"userID" validate:"required"`
}

type ManageRoleChange struct {
	UserID  uint                `json:"userId" validate:"required"`
	NewRole model.UserRole      `json:"newRole" validate:"required,oneof=user landlords"`
	Status  model.RequestStatus `json:"status"`
}

type RoleChangeRequestWithUser struct {
	ID       uint
	UserID   uint
	UserName string
	NewRole  model.UserRole
	Status   model.RequestStatus
}
