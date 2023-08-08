package main

import (
	"go-appointement/routes"
	"go-appointement/storage"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/kataras/iris/v12"
)

func main() {
	godotenv.Load()
	storage.InitializeDb()
	storage.InitializeRedis()
	app := iris.Default()

	app.Validator = validator.New()

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

	app.Listen(":8080")
}
