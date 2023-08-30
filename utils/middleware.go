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
		ctx.JSON(iris.Map{
			"MESSAGE": "ONLY Landlords and Admin can create Property",
			"STATUS":  iris.StatusForbidden,
		})
	}
}
