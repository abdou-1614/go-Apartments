package utils

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
)

func RoleMiddleware(roles ...string) iris.Handler {
	return func(ctx iris.Context) {
		//params := ctx.Params()
		/*id := params.Get("id")*/
		claims := jwt.Get(ctx).(*AccessToken)

		for _, role := range roles {
			if string(claims.ROLE) == role {
				ctx.Next()
				return
			}
		}
		response := map[string]interface{}{
			"MESSAGE": "Not Allowed to perform this action",
			"STATUS":  iris.StatusForbidden,
		}
		ctx.JSON(response)
	}
}
