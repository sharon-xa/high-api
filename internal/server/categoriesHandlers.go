package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sharon-xa/high-api/internal/database"
	"github.com/sharon-xa/high-api/internal/utils"
	"gorm.io/gorm"
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

type categoryReq struct {
	Name string `json:"name"`
}

func (s *Server) addCategory(c *gin.Context) {
	var req categoryReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		utils.Fail(c, utils.ErrBadRequest, nil)
		return
	}
	var category database.Category

	categoryName := capitalizeString(req.Name)
	category.Name = categoryName

	err = s.db.Create(&category).Error
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	utils.Created(c, "category created successfully", category)
}

func (s *Server) updateCategory(c *gin.Context) {
	categoryID := convParamToInt(c, "id")
	if categoryID == 0 {
		return
	}

	var req categoryReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		utils.Fail(c, utils.ErrBadRequest, nil)
		return
	}

	var category database.Category
	err = s.db.First(&category, categoryID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, utils.ErrNotFound, err)
			return
		}
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	categoryName := capitalizeString(req.Name)
	category.Name = categoryName

	err = s.db.Save(&category).Error
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	utils.Success(c, "category updated successfully", category)
}

func (s *Server) deleteCategory(c *gin.Context) {
	categoryID := convParamToInt(c, "id")
	if categoryID == 0 {
		return
	}

	results := s.db.Delete(&database.Category{}, categoryID)
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

	utils.Success(c, "category deleted successfully", nil)
}
