package server

import (
	"errors"
	"net/http"
	"strings"

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

type tagReq struct {
	Name string `json:"name"`
}

func (s *Server) updateTag(c *gin.Context) {
	var req tagReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		utils.Fail(c, utils.ErrBadRequest, err)
		return
	}

	tagID := convParamToInt(c, "id")
	if tagID == 0 {
		return
	}

	var tag database.Tag
	if err := s.db.First(&tag, tagID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, utils.ErrNotFound, err)
		} else {
			utils.Fail(c, utils.ErrInternal, err)
		}
		return
	}

	tag.Name = strings.ToLower(strings.TrimSpace(req.Name))

	if err := s.db.Save(&tag).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			utils.Fail(c, &utils.APIError{
				Code:    http.StatusConflict,
				Message: "Tag name must be unique",
			}, err)
			return
		}
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	utils.Success(c, "Tag updated successfully", tag)
}

func (s *Server) deleteTag(c *gin.Context) {
	tagID := convParamToInt(c, "id")
	if tagID == 0 {
		return
	}

	results := s.db.Delete(&database.Tag{}, tagID)
	if results.Error != nil {
		utils.Fail(c, utils.ErrInternal, results.Error)
		return
	}
	if results.RowsAffected == 0 {
		utils.Fail(
			c,
			&utils.APIError{Code: http.StatusNotFound, Message: "category not found"},
			nil,
		)
		return
	}

	utils.Success(c, "Tag deleted successfully", nil)
}
