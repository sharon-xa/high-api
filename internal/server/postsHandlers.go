package server

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sharon-xa/high-api/internal/database"
	"github.com/sharon-xa/high-api/internal/utils"
	"gorm.io/gorm"
)

type PostResponse struct {
	ID        uint          `json:"id"`
	Title     string        `json:"title"`
	Content   string        `json:"content"`
	Image     string        `json:"image"`
	CreatedAt time.Time     `json:"created_at"`
	User      UserBrief     `json:"user"`
	Category  CategoryBrief `json:"category"`
	Tags      []string      `json:"tags"`
}

type UserBrief struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

type CategoryBrief struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func (s *Server) getPost(c *gin.Context) {
	postId := convParamToInt(c, "id")
	if postId == 0 {
		return
	}

	var p database.Post
	if err := s.db.Preload("User").Preload("Category").Preload("Tags").First(&p, postId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, utils.ErrNotFound, err)
			return
		}

		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	tags := make([]string, len(p.Tags))
	for i, tag := range p.Tags {
		tags[i] = tag.Name
	}

	utils.Success(c, "", PostResponse{
		ID:        p.ID,
		Title:     p.Title,
		Content:   p.Content,
		Image:     p.Image,
		CreatedAt: p.CreatedAt,
		Tags:      tags,
		User: UserBrief{
			ID:    p.UserID,
			Name:  p.User.Name,
			Image: p.User.Image,
		},
		Category: CategoryBrief{
			ID:   p.Category.ID,
			Name: p.Category.Name,
		},
	})
}

type commentResponse struct {
	ID        uint              `json:"id"`
	Content   string            `json:"content"`
	CreatedAt time.Time         `json:"created_at"`
	User      publicUserSummary `json:"user"`
}

type publicUserSummary struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

func (s *Server) getCommentsOfPost(c *gin.Context) {
	postId := convParamToInt(c, "id")
	if postId == 0 {
		utils.Fail(c, utils.ErrBadRequest, nil)
		return
	}

	var comments []database.Comment
	err := s.db.Preload("User").
		Where("post_id = ?", postId).
		Order("created_at ASC").
		Find(&comments).
		Error
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	var response []commentResponse
	for _, cm := range comments {
		response = append(response, commentResponse{
			ID:        cm.ID,
			Content:   cm.Content,
			CreatedAt: cm.CreatedAt,
			User: publicUserSummary{
				ID:    cm.User.ID,
				Name:  cm.User.Name,
				Image: cm.User.Image,
			},
		})
	}
	if len(response) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	// TODO: we might want to limit the number of comments returned to the user
	// paginate the comments
	utils.Success(c, "", response)
}

func (s *Server) addPost(c *gin.Context) {
	claims := getAccessClaims(c)
	if claims == nil {
		return
	}

	userId, err := strconv.Atoi(claims.Subject)
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	categoryId := getRequiredFormFieldUInt(c, "categoryId")
	if categoryId == 0 {
		return
	}
	title := getRequiredFormField(c, "title")
	if title == "" {
		return
	}
	content := getRequiredFormField(c, "content")
	if content == "" {
		return
	}
	tagsStr := getRequiredFormField(c, "tags")
	if tagsStr == "" {
		return
	}

	// Split tags and trim whitespace
	rawTags := strings.Split(tagsStr, ",")
	var tags []database.Tag
	for _, tagName := range rawTags {
		tagName = strings.TrimSpace(tagName)
		tagName = strings.ToLower(tagName)
		if tagName == "" {
			continue
		}
		var tag database.Tag
		err := s.db.Where("name = ?", tagName).
			FirstOrCreate(&tag, database.Tag{Name: tagName}).
			Error
		if err != nil {
			utils.Fail(
				c,
				&utils.APIError{
					Code:    http.StatusInternalServerError,
					Message: "Failed to create or fetch tag",
				},
				err,
			)
			return
		}
		tags = append(tags, tag)
	}

	image, imageHeader, err := c.Request.FormFile("image")
	if err != nil {
		utils.Fail(
			c,
			&utils.APIError{Code: http.StatusBadRequest, Message: "image is required"},
			err,
		)
		return
	}
	defer image.Close()
	url, err := s.s3.UploadImage(c, image, imageHeader)
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	defer image.Close()
	p := database.Post{
		UserID:     uint(userId),
		CategoryID: categoryId,
		Title:      title,
		Content:    content,
		Tags:       tags,
		Image:      url,
	}
	if err := s.db.Create(&p).Error; err != nil {
		utils.Fail(c, &utils.APIError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create post",
		}, err)
		return
	}

	if err := s.db.Preload("User").Preload("Category").First(&p, p.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load related data"})
		return
	}

	CreatedTags := make([]string, len(p.Tags))
	for i, tag := range p.Tags {
		CreatedTags[i] = tag.Name
	}

	utils.Created(c, "post created successfully", PostResponse{
		ID:        p.ID,
		Title:     p.Title,
		Content:   p.Content,
		Image:     p.Image,
		CreatedAt: p.CreatedAt,
		Tags:      CreatedTags,
		User: UserBrief{
			ID:    p.UserID,
			Name:  p.User.Name,
			Image: p.User.Image,
		},
		Category: CategoryBrief{
			ID:   p.Category.ID,
			Name: p.Category.Name,
		},
	})
}

type updatePostRequest struct {
	Title      string `json:"title"      binding:"required"`
	Content    string `json:"content"    binding:"required"`
	CategoryID uint   `json:"categoryId" binding:"required"`
	Tags       string `json:"tags"       binding:"required"`
}

func (s *Server) updatePost(c *gin.Context) {
	var req updatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.ErrBadRequest, err)
		return
	}

	postId := convParamToInt(c, "id")
	if postId == 0 {
		utils.Fail(c, utils.ErrBadRequest, errors.New("invalid post ID"))
		return
	}

	claims := getAccessClaims(c)
	if claims == nil {
		return
	}

	var post database.Post
	if err := s.db.Preload("Tags").First(&post, postId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, utils.ErrNotFound, err)
		} else {
			utils.Fail(c, utils.ErrInternal, err)
		}
		return
	}

	userId, err := strconv.Atoi(claims.Subject)
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}
	if post.UserID != uint(userId) {
		utils.Fail(
			c,
			&utils.APIError{
				Code:    http.StatusForbidden,
				Message: "you're not allowed to delete this post",
			},
			nil,
		)
		return
	}

	post.Title = req.Title
	post.Content = req.Content
	post.CategoryID = req.CategoryID

	rawTags := strings.Split(req.Tags, ",")
	var tags []database.Tag
	for _, tagName := range rawTags {
		tagName = strings.ToLower(strings.TrimSpace(tagName))
		if tagName == "" {
			continue
		}
		var tag database.Tag
		if err := s.db.Where("name = ?", tagName).FirstOrCreate(&tag, database.Tag{Name: tagName}).Error; err != nil {
			utils.Fail(c, utils.ErrInternal, err)
			return
		}
		tags = append(tags, tag)
	}
	s.db.Model(&post).Association("Tags").Replace(&tags)

	if err := s.db.Save(&post).Error; err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	utils.Success(c, "Post updated successfully", nil)
}

func (s *Server) deletePost(c *gin.Context) {
	postId := convParamToInt(c, "id")
	if postId == 0 {
		utils.Fail(c, utils.ErrBadRequest, nil)
		return
	}

	claims := getAccessClaims(c)
	if claims == nil {
		return
	}

	var p database.Post
	err := s.db.First(&p, postId).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(
				c,
				&utils.APIError{Code: http.StatusNotFound, Message: "post doesn't exist"},
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
	if p.UserID != uint(userId) && claims.Role != "admin" {
		utils.Fail(
			c,
			&utils.APIError{
				Code:    http.StatusForbidden,
				Message: "you're not allowed to delete this post",
			},
			nil,
		)
		return
	}

	results := s.db.Delete(&database.Post{}, p.ID)
	if results.Error != nil {
		utils.Fail(c, utils.ErrInternal, results.Error)
		return
	}

	if results.RowsAffected == 0 {
		utils.Fail(
			c,
			&utils.APIError{Code: http.StatusNotFound, Message: "post not found"},
			nil,
		)
		return
	}

	utils.Success(c, "post deleted successfully", nil)
}
