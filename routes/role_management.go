package routes

import (
	"go-appointement/model"
	"go-appointement/storage"
	"go-appointement/utils"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
)

// Submit Change Role.
//
// This endpoint allows users to submit to change role.
// @Summary Submit Change Role Request.
// @Description Submit user request to change role.
// @Tags Users
// @Security JWT
// @Accept json
// @Produce json
// @Param request body RoleChangeRequest true "User Submit"
// @Success 200  "THE REQUEST HAS BEEN SUBMITED SUCCESSFULLY"
// @Failure 400 "Invalid input"
// @Failure 500 "You can only request a role change for yourself"
// @Router /submit-role-change [post]
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

// ManageRoleChangeRequests
// @Summary Manage All users requests
// @Description Retrieves All user requests to change role.
// @Accept json
// @Produce json
// @Success 200 {array} RoleChangeRequestWithUser
// @Tags Users
// @Security JWT
// @Failure 500 "Internal Server Error"
// @Failure 403 "Only admins can manage role change requests"
// @Router /manage-role-requests [get]
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

// @Summary Accept a role change request
// @Description Accepts a role change request for an admin user
// @ID accept-role-change-request
// @Accept json
// @Tags Users
// @Security JWT
// @Produce json
// @Param id path int true "Request ID"
// @Success 200 {object} map[string]interface{} "Success response"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Not Found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /accept-role-change-request/{id} [put]
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

// @Summary Reject a role change request
// @Description Rejects a role change request for an admin user
// @ID reject-role-change-request
// @Accept json
// @Tags Users
// @Security JWT
// @Produce json
// @Param id path int true "Request ID"
// @Success 200 {object} map[string]interface{} "Success response"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Not Found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /reject-role-request/{id} [put]
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
