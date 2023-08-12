package utils

import (
	"strconv"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
)

func RoleMiddleware(roles ...string) iris.Handler {
	return func(ctx iris.Context) {
		claims := jwt.Get(ctx).(*AccessToken)

		params := ctx.Params()
		id := params.Get("id")

		if strconv.FormatUint(uint64(claims.ID), 10) != id {
			ctx.StatusCode(iris.StatusForbidden)
			return
		}

		for _, role := range roles {
			if claims.ROLE == role {
				ctx.Next()
				return
			}
		}
		ctx.StatusCode(iris.StatusForbidden)
	}
}
