package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sharon-xa/high-api/internal/auth"
	"github.com/sharon-xa/high-api/internal/utils"
)

func Admin(accessTokenSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := auth.GetAccessClaimsFromAuthHeader(c, accessTokenSecret)

		if claims == nil {
			return
		}

		if claims.Role != "admin" {
			utils.FailAndAbort(
				c,
				utils.NewAPIError(http.StatusForbidden, utils.ErrRoleNotAllowed.Error()),
				nil,
			)
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}

func User(accessTokenSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := auth.GetAccessClaimsFromAuthHeader(c, accessTokenSecret)

		if claims == nil {
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}
