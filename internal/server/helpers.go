package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sharon-xa/high-api/internal/auth"
	"github.com/sharon-xa/high-api/internal/config"
	"github.com/sharon-xa/high-api/internal/utils"
)

func convParamToInt(c *gin.Context, param string) uint {
	paramStr := c.Param(param)
	paramInt, err := strconv.Atoi(paramStr)
	if err != nil || paramInt <= 0 {
		utils.Fail(c, utils.ErrBadRequest, err)
		return 0
	}

	return uint(paramInt)
}

func convStrToUInt(c *gin.Context, numAsStr string, fieldName string) uint {
	val, err := strconv.Atoi(numAsStr)
	if err != nil || val < 1 {
		utils.Fail(
			c,
			&utils.APIError{Code: http.StatusBadRequest, Message: "Invalid " + fieldName},
			err,
		)
		return 0
	}

	return uint(val)
}

func getHeader(c *gin.Context, key string) string {
	header := strings.TrimSpace(c.GetHeader(key))
	if header == "" {
		utils.Fail(
			c,
			utils.ErrHeaderMissing(key),
			nil,
		)
		return ""
	}
	return header
}

func getRequiredFormField(c *gin.Context, formFieldName string) string {
	field := c.PostForm(formFieldName)
	if field == "" {
		utils.Fail(
			c,
			&utils.APIError{
				Code:    http.StatusBadRequest,
				Message: formFieldName + " field is required",
			},
			fmt.Errorf("the %s field is empty", formFieldName),
		)
		return ""
	}

	return field
}

func getRequiredFormFieldUInt(c *gin.Context, formFieldName string) uint {
	field := getRequiredFormField(c, formFieldName)
	if field == "" {
		return 0
	}

	fieldInt := convStrToUInt(c, field, formFieldName)
	if fieldInt == 0 {
		return 0
	}

	return fieldInt
}

func setCookie(c *gin.Context, cookieName, cookieVal string, expTimeInSec int, env *config.Env) {
	var secure bool = true
	if env.Environment == "dev" {
		secure = false
	}

	c.SetCookie(
		cookieName,
		cookieVal,
		expTimeInSec,
		"/",
		env.ApiDomain,
		secure,
		true,
	)
}

func getAccessClaims(c *gin.Context) *auth.AccessClaims {
	claimsInterface, exists := c.Get("claims")
	if !exists {
		utils.Fail(c, utils.ErrUnauthorized, nil)
		return nil
	}

	claims, ok := claimsInterface.(*auth.AccessClaims)
	if !ok {
		utils.Fail(c, utils.ErrUnauthorized, nil)
		return nil
	}
	return claims
}
