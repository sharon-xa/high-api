package server

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/sharon-xa/high-api/internal/database"
	"github.com/sharon-xa/high-api/internal/utils"
	"gorm.io/gorm"
)

type publicUserResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
	Bio   string `json:"bio"`
}

func (s *Server) getUserPublic(c *gin.Context) {
	userId := convParamToInt(c, "id")
	if userId == 0 {
		utils.Fail(c, utils.ErrBadRequest, nil)
		return
	}

	var user publicUserResponse
	err := s.db.Model(&database.User{}).
		Select("id", "name", "image", "bio").
		Where("id = ?", userId).
		First(&user).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, utils.ErrNotFound, err)
			return
		}
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	utils.Success(c, "", user)
}
