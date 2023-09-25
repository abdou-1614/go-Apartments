package routes

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/kataras/iris/v12"
)

// AutoComplete performs location autocomplete using a third-party API.
// @Summary Location Autocomplete
// @Description Get location suggestions based on user input.
// @Tags Location
// @Accept json
// @Produce json
// @Param location query string true "Location input for autocomplete"
// @Param limit query int false "Limit the number of suggestions (default: 10)"
// @Success 200 {array} string "An array of location suggestions"
// @Router /location/autocomplete [get]
func AutoComplete(ctx iris.Context) {
	limit := "10"
	location := ctx.URLParam("location")
	limitQuery := ctx.URLParam("limit")

	if limitQuery != "" {
		limit = limitQuery
	}

	apiKey := os.Getenv("LOCATION-API")

	url := "https://api.locationiq.com/v1/autocomplete.php?key=" + apiKey + "&q=" + location + "&limit=" + limit

	fetchLocation(url, ctx)
}

func Search(ctx iris.Context) {
	location := ctx.URLParam("location")
	apiKey := os.Getenv("LOCATION-API")

	url := "https://api.locationiq.com/v1/autocomplete.php?key=" + apiKey + "&q=" + location + "&format=json"

	fetchLocation(url, ctx)
}

func fetchLocation(url string, ctx iris.Context) {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", url, nil)

	res, locationErr := client.Do(req)

	if locationErr != nil {
		ctx.StopWithProblem(iris.StatusInternalServerError, iris.NewProblem().
			Title("Internal Server Error").
			Detail("Internal Server Error"),
		)
		return
	}

	defer res.Body.Close()

	body, bodyErr := ioutil.ReadAll(res.Body)

	if bodyErr != nil {
		ctx.StopWithProblem(iris.StatusInternalServerError, iris.NewProblem().
			Title("Internal Server Error").
			Detail("Internal Server Error"),
		)
		return
	}

	var obgJson []map[string]interface{}

	jsonErr := json.Unmarshal(body, &obgJson)

	if jsonErr != nil {
		ctx.StopWithProblem(iris.StatusInternalServerError, iris.NewProblem().
			Title("Internal Server Error").
			Detail("Internal Server Error"),
		)

		ctx.JSON(obgJson)
	}
}
