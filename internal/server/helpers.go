package server

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sharon-xa/high-api/internal/utils"
)

func convParamToInt(c *gin.Context, param string) int32 {
	paramStr := c.Param(param)
	paramInt, err := strconv.Atoi(paramStr)
	if err != nil || paramInt <= 0 {
		utils.Fail(c, utils.ErrBadRequest, err)
		return 0
	}

	return int32(paramInt)
}

// this func asumes that the string isn't ""
func convStrToInt32(c *gin.Context, str string) int32 {
	convertedInt, err := strconv.Atoi(str)
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return 0
	}

	return int32(convertedInt)
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
