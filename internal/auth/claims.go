package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sharon-xa/high-api/internal/utils"
)

type AccessClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func GetAccessClaimsFromAuthHeader(c *gin.Context, accessTokenSecret string) *AccessClaims {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		utils.FailAndAbort(
			c,
			utils.NewAPIError(http.StatusUnauthorized, "Authorization header missing"),
			nil,
		)
		return nil
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	claims, err := ParseAccessToken(tokenString, accessTokenSecret)
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			utils.FailAndAbort(c, utils.ErrTokenExpired, nil)
		case errors.Is(err, jwt.ErrTokenInvalidClaims):
			utils.FailAndAbort(c, utils.ErrTokenInvalidClaims, nil)
		case errors.Is(err, utils.ErrParsingToken):
			utils.FailAndAbort(c, utils.ErrParsingToken, err)
		case errors.Is(err, utils.ErrInvalidToken):
			utils.FailAndAbort(c, utils.ErrInvalidToken, err)
		default:
			utils.FailAndAbort(c, utils.ErrUnauthorized, err)
		}
		return nil
	}

	return claims
}
