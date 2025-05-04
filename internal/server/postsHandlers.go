package server

import (
	"errors"
	"net/http"
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

func (s *Server) addPost(c *gin.Context) {
	userId := getRequiredFormFieldUInt(c, "userId")
	if userId == 0 {
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
		UserID:     userId,
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
