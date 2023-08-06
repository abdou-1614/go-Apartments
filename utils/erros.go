package utils

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/kataras/iris/v12"
)

func CreateError(statusCode int, title string, details string, ctx iris.Context) {
	ctx.StopWithProblem(statusCode, iris.NewProblem().Title(title).Detail(details))
}

func CreateInternalServerError(ctx iris.Context) {
	CreateError(iris.StatusInternalServerError, "Internal Server Error", "Internal Server Error", ctx)
}

func HandleValidationError(err error, ctx iris.Context) {
	if errs, ok := err.(validator.ValidationErrors); ok {
		validationErrors := wrapeValidationError(errs)
		fmt.Println("ValidationError", validationErrors)
		ctx.StopWithProblem(
			iris.StatusBadRequest,
			iris.NewProblem().Title("Validation Error").Detail("One or More Field faild To Be Validated").Key("Errors", validationErrors),
		)
		return
	}
	CreateInternalServerError(ctx)
}

func wrapeValidationError(errs validator.ValidationErrors) []validationError {
	validationErrors := make([]validationError, 0, len(errs))
	for _, validationErr := range errs {
		validationErrors = append(validationErrors, validationError{
			ActualTag: validationErr.ActualTag(),
			NameSpace: validationErr.Namespace(),
			Kind:      validationErr.Kind().String(),
			Type:      validationErr.Type().String(),
			Value:     validationErr.Type().String(),
			Param:     validationErr.Param(),
		})
	}
	return validationErrors
}

type validationError struct {
	ActualTag string `json:"tag"`
	NameSpace string `json:"namespace"`
	Kind      string `json:"kind"`
	Type      string `json:"type"`
	Value     string `json:"value"`
	Param     string `json:"param"`
}
