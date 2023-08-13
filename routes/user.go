package routes

import (
	"go-appointement/model"
	"strings"

	"go-appointement/storage"
	"go-appointement/utils"

	"github.com/kataras/iris/v12"
	"golang.org/x/crypto/bcrypt"
)

func Login(ctx iris.Context) {
	var userInput LoginUserInput

	err := ctx.JSON(&userInput)

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

	if userExist == false {
		utils.CreateError(iris.StatusUnauthorized, "Credentials Error", "Invalid Email Or Password", ctx)
		return
	}

	if existUser.SocialLogin == true {
		utils.CreateError(iris.StatusUnauthorized, "Credentials Error", "Social Login Account", ctx)
		return
	}

	passwordErr := bcrypt.CompareHashAndPassword([]byte(existUser.Password), []byte(userInput.Password))

	if passwordErr != nil {
		utils.CreateError(iris.StatusUnauthorized, "Credentials Error", "Invalid Email Or Password", ctx)
		return
	}
	ctx.JSON(iris.Map{
		"ID":        existUser.ID,
		"FIRSTNAME": existUser.FirstName,
		"LASTNAME":  existUser.LastName,
		"EMAIL":     existUser.Email,
	})
}

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

	if userExist == true {
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

func HandleUserExits(user *model.User, email string) (exists bool, err error) {
	userExistQuery := storage.DB.Where("email = ?", strings.ToLower(email)).Limit(1).Find(&user)

	if userExistQuery.Error != nil {
		return false, userExistQuery.Error
	}

	userExist := userExistQuery.RowsAffected > 0

	if userExist == true {
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
