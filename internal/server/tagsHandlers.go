package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sharon-xa/high-api/internal/database"
	"github.com/sharon-xa/high-api/internal/utils"
	"gorm.io/gorm"
)

func (s *Server) getAllTags(c *gin.Context) {
	var t []database.Tag
	err := s.db.Find(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(
				c,
				&utils.APIError{Code: http.StatusNoContent, Message: "no tags found"},
				err,
			)
			return
		}
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	utils.Success(c, "", t)
}
