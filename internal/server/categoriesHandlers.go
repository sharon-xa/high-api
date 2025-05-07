package server

import (
	"github.com/gin-gonic/gin"
	"github.com/sharon-xa/high-api/internal/database"
	"github.com/sharon-xa/high-api/internal/utils"
)

func (s *Server) getCategories(c *gin.Context) {
	var categories []database.Category

	err := s.db.Select("id", "name").Find(&categories).Error
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	utils.Success(c, "", categories)
}
