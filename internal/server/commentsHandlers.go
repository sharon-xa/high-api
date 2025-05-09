package server

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sharon-xa/high-api/internal/database"
	"github.com/sharon-xa/high-api/internal/utils"
	"gorm.io/gorm"
)

type commentReq struct {
	Content string `json:"content" binding:"required"`
}

func (s *Server) addComment(c *gin.Context) {
	claims := getAccessClaims(c)
	if claims == nil {
		return
	}

	var req commentReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		utils.Fail(c, utils.ErrBadRequest, err)
		return
	}

	postId := convParamToInt(c, "id")
	if postId == 0 {
		return
	}

	userId, err := strconv.Atoi(claims.Subject)
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	var comment database.Comment

	comment.Content = req.Content
	comment.PostID = postId
	comment.UserID = uint(userId)

	err = s.db.Create(&comment).Error
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	var user database.User
	if err := s.db.Select("id", "name", "image").First(&user, comment.UserID).Error; err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	resp := commentResponse{
		ID:        comment.ID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
		User: publicUserSummary{
			ID:    user.ID,
			Name:  user.Name,
			Image: user.Image,
		},
	}

	utils.Success(c, "comment created successfully", resp)
}

func (s *Server) updateComment(c *gin.Context) {
	claims := getAccessClaims(c)
	if claims == nil {
		return
	}

	var req commentReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		utils.Fail(c, utils.ErrBadRequest, nil)
		return
	}

	commentId := convParamToInt(c, "id")
	if commentId == 0 {
		return
	}

	var comment database.Comment
	err = s.db.First(&comment, commentId).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(
				c,
				&utils.APIError{Code: http.StatusNotFound, Message: "comment not found"},
				err,
			)
			return
		}
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	userId, err := strconv.Atoi(claims.Subject)
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}
	if comment.UserID != uint(userId) {
		utils.Fail(
			c,
			&utils.APIError{
				Code:    http.StatusForbidden,
				Message: "you're not allowed to change this comment",
			},
			err,
		)
		return
	}

	comment.Content = req.Content

	if err := s.db.Save(&comment).Error; err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	var user database.User
	if err := s.db.Select("id", "name", "image").First(&user, comment.UserID).Error; err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	resp := commentResponse{
		ID:        comment.ID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
		User: publicUserSummary{
			ID:    user.ID,
			Name:  user.Name,
			Image: user.Image,
		},
	}

	utils.Success(c, "comment updated successfully", resp)
}

func (s *Server) deleteComment(c *gin.Context) {
	claims := getAccessClaims(c)
	if claims == nil {
		return
	}

	commentId := convParamToInt(c, "id")
	if commentId == 0 {
		return
	}

	var comment database.Comment
	if err := s.db.First(&comment, commentId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(
				c,
				&utils.APIError{Code: http.StatusNotFound, Message: "comment not found"},
				err,
			)
			return
		}
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	userId, err := strconv.Atoi(claims.Subject)
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	if comment.UserID != uint(userId) && claims.Role != "admin" {
		utils.Fail(
			c,
			&utils.APIError{
				Code:    http.StatusForbidden,
				Message: "you're not allowed to remove this comment",
			},
			nil,
		)
		return
	}

	if err := s.db.Delete(&comment).Error; err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	utils.Success(c, "comment deleted successfully", nil)
}
