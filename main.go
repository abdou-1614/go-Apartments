package main

import (
	"go-appointement/routes"
	"go-appointement/storage"

	"github.com/joho/godotenv"
	"github.com/kataras/iris/v12"
)

func main() {
	godotenv.Load()
	storage.InitializeDb()
	app := iris.Default()

	location := app.Party("api/location")

	{
		location.Get("/autocomplete", routes.AutoComplete)
		location.Get("/search", routes.Search)
	}

	user := app.Party("api/user")

	{
		user.Post("/create", routes.Register)
	}

	app.Listen(":8080")
}
