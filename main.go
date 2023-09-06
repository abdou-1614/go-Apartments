package main

import (
	"go-appointement/model"
	"go-appointement/routes"
	"go-appointement/storage"
	"go-appointement/utils"
	"os"

	_ "go-appointement/docs"

	"github.com/go-playground/validator/v10"
	"github.com/iris-contrib/swagger"
	"github.com/iris-contrib/swagger/swaggerFiles"
	"github.com/joho/godotenv"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
)

// @title Swagger Example APARTEMENTS
// @version 1.0
// @description This is a sample server Petstore server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
// @host localhost:8080
// @BasePath /api

func main() {
	godotenv.Load()
	storage.InitializeDb()
	storage.InitializeRedis()
	storage.IntialzeS3()
	app := iris.Default()

	app.Validator = validator.New()

	resetTokenVerifyer := jwt.NewVerifier(jwt.HS256, []byte(os.Getenv("EMAIL_SECRET_TOKEN")))
	resetTokenVerifyer.WithDefaultBlocklist()
	resetTokenMiddleware := resetTokenVerifyer.Verify(func() interface{} {
		return new(utils.ForgetPasswordToken)
	})

	accessTokenVerifier := jwt.NewVerifier(jwt.HS256, []byte(os.Getenv("ACCESS_TOKEN_SECRET")))
	accessTokenVerifier.WithDefaultBlocklist()
	accessTokenVerifierMiddleware := accessTokenVerifier.Verify(func() interface{} {
		return new(utils.AccessToken)
	})

	refreshTokenVerifier := jwt.NewVerifier(jwt.HS256, []byte(os.Getenv("REFRESH_TOKEN_SECRET")))
	refreshTokenVerifier.WithDefaultBlocklist()
	refreshTokenVerifierMiddleware := refreshTokenVerifier.Verify(func() interface{} {
		return new(jwt.Claims)
	})

	refreshTokenVerifier.Extractors = append(refreshTokenVerifier.Extractors, func(ctx iris.Context) string {
		var tokenInput utils.RefreshTokenInput
		err := ctx.JSON(&tokenInput)
		if err != nil {
			return ""
		}

		return tokenInput.RefreshToken
	})

	location := app.Party("api/location")

	{
		location.Get("/autocomplete", routes.AutoComplete)
		location.Get("/search", accessTokenVerifierMiddleware, routes.Search)
	}

	user := app.Party("api/user")

	{
		user.Post("/register", routes.Register)
		user.Post("/login", routes.Login)
		user.Post("/forgetpassword", routes.ForgetPassword)
		user.Post("/restpassword", resetTokenMiddleware, routes.RestPassword)
	}

	property := app.Party("api/property")

	{
		property.Post("/create", accessTokenVerifierMiddleware, utils.RoleMiddleware(string(model.RoleLandlords), string(model.RoleAdmin)), routes.CreateProperty)
		property.Get("/{id}", routes.GetProperty)
		property.Get("/all", routes.GetAllProperty)
		property.Get("/top-rated", routes.GetTopRatedPropert)
		property.Delete("/delete/{id}", accessTokenVerifierMiddleware, routes.DeleteProperty)
		property.Patch("/update/{id}", utils.RoleMiddleware(string(model.RoleLandlords), string(model.RoleAdmin)), routes.UpdateProperty)
	}

	app.Post("/api/refresh", refreshTokenVerifierMiddleware, utils.RefreshToken)

	swaggerUI := swagger.Handler(swaggerFiles.Handler)

	// Register on http://localhost:8080/swagger
	app.Get("/swagger", swaggerUI)
	app.Get("/swagger/{any:path}", swaggerUI)

	app.Listen(":8080")
}
