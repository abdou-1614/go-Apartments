package main

import (
	"go-appointement/routes"
	"go-appointement/storage"
	"go-appointement/utils"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
)

func main() {
	godotenv.Load()
	storage.InitializeDb()
	storage.InitializeRedis()
	app := iris.Default()

	app.Validator = validator.New()

	accessTokenVerifier := jwt.NewVerifier(jwt.HS256, os.Getenv("ACCESS_TOKEN_SECRET"))
	accessTokenVerifier.WithDefaultBlocklist()
	accessTokenVerifierMiddleware := accessTokenVerifier.Verify(func() interface{} {
		return new(utils.AccessToken)
	})

	refreshTokenVerifier := jwt.NewVerifier(jwt.HS256, os.Getenv("REFRESH_TOKEN_SECRET"))
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
		location.Get("/search", routes.Search)
	}

	user := app.Party("api/user")

	{
		user.Post("/register", routes.Register)
		user.Post("/login", routes.Login)
	}

	app.Post("/api/refresh", refreshTokenVerifierMiddleware, utils.RefreshToken)

	app.Listen(":8080")
}
