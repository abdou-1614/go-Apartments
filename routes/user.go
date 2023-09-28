package routes

import (
	"go-appointement/model"
	"strings"

	"go-appointement/storage"
	"go-appointement/utils"

	"github.com/kataras/iris/v12"
	jsonWt "github.com/kataras/iris/v12/middleware/jwt"
	"golang.org/x/crypto/bcrypt"
)

func Login(ctx iris.Context) {
	var userInput LoginUserInput

	err := ctx.ReadJSON(&userInput)

	if err != nil {
		utils.HandleValidationError(err, ctx)
		return
	}

	var existUser model.User

	userExist, userExistErr := HandleUserExits(&existUser, userInput.Email)

	if userExistErr != nil {
		utils.CreateInternalServerError(ctx)
		return
	}

	if !userExist {
		utils.CreateError(iris.StatusUnauthorized, "Credentials Error", "Invalid Email Or Password", ctx)
		return
	}

	if existUser.SocialLogin {
		utils.CreateError(iris.StatusUnauthorized, "Credentials Error", "Social Login Account", ctx)
		return
	}

	passwordErr := bcrypt.CompareHashAndPassword([]byte(existUser.Password), []byte(userInput.Password))

	if passwordErr != nil {
		utils.CreateError(iris.StatusUnauthorized, "Credentials Error", "Invalid Email Or Password", ctx)
		return
	}

	returnUser(existUser, ctx)
}

// Register registers a new user.
//
// This endpoint allows users to register a new account.
//
// @Summary Register a new user
// @Description Register a new user with the provided information.
// @Tags Users
// @Accept json
// @Produce json
// @Param request body RegisterUser true "User registration details"
// @Success 201 {object} RegisterUser "User registered successfully"
// @Failure 400 "Invalid input"
// @Failure 409 "User already exists"
// @Router /user/register [post]
func Register(ctx iris.Context) {
	var userInput RegisterUser

	err := ctx.ReadJSON(&userInput)

	if err != nil {
		utils.HandleValidationError(err, ctx)
		return
	}

	var existUser model.User

	userExist, userExistErr := HandleUserExits(&existUser, userInput.Email)

	if userExistErr != nil {
		utils.CreateInternalServerError(ctx)
		return
	}

	if userExist {
		utils.CreateError(iris.StatusConflict, "Conflict", "User Already Exist", ctx)
		return
	}

	hashedPassword, hashedErr := HashPassword(userInput.Password)

	if hashedErr != nil {
		utils.CreateInternalServerError(ctx)
		return
	}

	existUser = model.User{
		FirstName:   userInput.FirstName,
		LastName:    userInput.LastName,
		Email:       strings.ToLower(userInput.Email),
		Password:    hashedPassword,
		SocialLogin: false,
		Role:        model.RoleUser,
	}

	storage.DB.Create(&existUser)

	ctx.JSON(iris.Map{
		"ID":        existUser.ID,
		"FIRSTNAME": existUser.FirstName,
		"LASTNAME":  existUser.LastName,
		"EMAIL":     existUser.Email,
		"ROLE":      existUser.Role,
	})

}

// ForgetPassword sends a password reset email to the user.
// @Summary Send password reset email
// @Description Sends a password reset email to the user's registered email address.
// @Tags Users
// @Accept json
// @Produce json
// @Param input body EmailRegisteredInput true "User's registered email address"
// @Success 200 {object} map[string]bool "Email sent successfully"
// @Failure 400 "Invalid Email" Example({"message": "Invalid Email"})
// @Failure 401 "Social Login Account" Example({"message": "Social Login Account"})
// @Failure 500 "Internal Server Error" Example({"message": "Internal Server Error"})
// @Router /user/forget-password [post]
func ForgetPassword(ctx iris.Context) {
	var emailInput EmailRegisteredInput
	err := ctx.ReadJSON(&emailInput)

	if err != nil {
		utils.HandleValidationError(err, ctx)
		return
	}

	var userModel model.User

	userExist, userExistErr := HandleUserExits(&userModel, emailInput.Email)

	if userExistErr != nil {
		utils.CreateInternalServerError(ctx)
		return
	}

	if !userExist {
		utils.CreateError(iris.StatusBadRequest, "Invalid Credentials", "Invalid Email", ctx)
		return
	}

	if userExist {
		if userModel.SocialLogin {
			utils.CreateError(iris.StatusBadRequest, "Credentials Error", "Social Login Account", ctx)
			return
		}
		link := "expo://localhost:19000/../resetpassword"
		token, tokenErr := utils.CreateForgetPasswordToken(userModel.ID, userModel.Email)

		if tokenErr != nil {
			utils.CreateInternalServerError(ctx)
			return
		}

		link += token

		subject := "Forget Your Password "

		html := `
			<p>Its Look Like You Forget Your Password.
			please click in link below to rest it.
			please rest your password within 10 minutes. otherwise you will have to repeat this
			process.<a href=` + link + `>click Here to rest password</a>
			</p>
		`
		emailSent, emailSentErr := utils.SendMail(userModel.Email, subject, html)

		if emailSentErr != nil {
			utils.CreateInternalServerError(ctx)
			return
		}

		if emailSent {
			ctx.JSON(iris.Map{
				"emailSent": true,
			})
			return
		}

		ctx.JSON(iris.Map{"emailSent": false})
	}
}

func RestPassword(ctx iris.Context) {
	var passwordInput RestPaswordInput
	err := ctx.ReadJSON(&passwordInput)

	if err != nil {
		utils.HandleValidationError(err, ctx)
		return
	}

	hashedPassword, hashedPasswordErr := HashPassword(passwordInput.Password)

	if hashedPasswordErr != nil {
		utils.CreateInternalServerError(ctx)
		return
	}

	claims := jsonWt.Get(ctx).(*utils.ForgetPasswordToken)
	var user model.User

	storage.DB.Model(&user).Where("id = ?", claims.ID).Update("password", hashedPassword)

	ctx.JSON(iris.Map{"passwordRest": true})
}

func HandleUserExits(user *model.User, email string) (exists bool, err error) {
	userExistQuery := storage.DB.Where("email = ?", strings.ToLower(email)).Limit(1).Find(&user)

	if userExistQuery.Error != nil {
		return false, userExistQuery.Error
	}

	userExist := userExistQuery.RowsAffected > 0

	if userExist {
		return true, nil
	}

	return false, nil
}

func HashPassword(password string) (hashedPassword string, err error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func returnUser(user model.User, ctx iris.Context) {
	tokenPair, tokenPairErr := utils.CreateToken(user.ID, user.Role)

	if tokenPairErr != nil {
		utils.CreateInternalServerError(ctx)
		return
	}

	ctx.JSON(iris.Map{
		"ID":           user.ID,
		"FIRSTNAME":    user.FirstName,
		"LASTNAME":     user.LastName,
		"EMAIL":        user.Email,
		"ROLE":         user.Role,
		"accessToken":  string(tokenPair.AccessToken),
		"refreshToken": string(tokenPair.RefreshToken),
	})
}

type EmailRegisteredInput struct {
	Email string `json:"email" validate:"required"`
}

type RegisterUser struct {
	FirstName string         `json:"firstName" validate:"required,max=265"`
	LastName  string         `json:"lastName" validate:"required,max=265"`
	Email     string         `json:"email" validate:"email,required,max=265"`
	Password  string         `json:"password" validate:"required,min=6,max=265"`
	Role      model.UserRole `json:"role" validate:"required" oneof:"user admin guest landlords"`
}

type LoginUserInput struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type RestPaswordInput struct {
	Password string `json:"password" validate:"required"`
}
