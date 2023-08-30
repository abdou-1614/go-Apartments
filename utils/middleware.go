package utils

import (
	"strconv"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
)

func RoleMiddleware(roles ...string) iris.Handler {
	return func(ctx iris.Context) {
		params := ctx.Params()
		id := params.Get("id")
		claims := jwt.Get(ctx).(*AccessToken)

		if strconv.FormatUint(uint64(claims.ID), 10) != id {
			ctx.JSON(iris.Map{
				"Message": "Invalid ID",
				"STATUS":  iris.StatusUnauthorized,
			})
			return
		}

		for _, role := range roles {
			if string(claims.ROLE) == role {
				ctx.Next()
				return
			}
		}
		ctx.StatusCode(iris.StatusForbidden)
	}
}
