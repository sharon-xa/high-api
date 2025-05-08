package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sharon-xa/high-api/internal/database"
	"github.com/sharon-xa/high-api/internal/utils"
	"gorm.io/gorm"
)

type adminUserResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Gender    string    `json:"gender"`
	Image     string    `json:"image"`
	Bio       string    `json:"bio"`
	Role      string    `json:"role"`
	Verified  bool      `json:"verified"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Server) getAllUsers(c *gin.Context) {
	var users []database.User

	err := s.db.Find(&users).Error
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	var response []adminUserResponse
	for _, u := range users {
		response = append(response, adminUserResponse{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			Gender:    u.Gender,
			Image:     u.Image,
			Bio:       u.Bio,
			Role:      u.Role,
			Verified:  u.Verified,
			CreatedAt: u.CreatedAt,
		})
	}

	// TODO: the users needs to be paginated
	utils.Success(c, "", response)
}

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

type userResponse struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Gender string `json:"gender"`
	Image  string `json:"image"`
	Bio    string `json:"bio"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

func (s *Server) getUser(c *gin.Context) {
	claims := getAccessClaims(c)
	if claims == nil {
		return
	}

	var u database.User
	err := s.db.First(&u, claims.Subject).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, utils.ErrNotFound, err)
			return
		}
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	response := userResponse{
		ID:     u.ID,
		Name:   u.Name,
		Email:  u.Email,
		Image:  u.Image,
		Bio:    u.Bio,
		Gender: u.Gender,
		Role:   u.Role,
	}

	utils.Success(c, "", response)
}

type updateUserReq struct {
	Name   string `json:"name"`
	Gender string `json:"gender"`
	Bio    string `json:"bio"`
}

func (s *Server) updateUser(c *gin.Context) {
	claims := getAccessClaims(c)
	if claims == nil {
		return
	}

	var req updateUserReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		utils.Fail(c, utils.ErrBadRequest, err)
		return
	}

	if req.Name == "" || len(req.Name) > 50 {
		utils.Fail(c, &utils.APIError{Code: http.StatusBadRequest, Message: "invalid name"}, nil)
		return
	}

	if req.Gender == "" || len(req.Gender) > 50 {
		utils.Fail(c, &utils.APIError{Code: http.StatusBadRequest, Message: "invalid gender"}, nil)
		return
	}

	var user database.User
	if err := s.db.First(&user, claims.Subject).Error; err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	user.Name = req.Name
	user.Bio = req.Bio
	user.Gender = req.Gender

	err = s.db.Save(&user).Error
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	utils.Success(c, "profile updated successfully", nil)
}

func (s *Server) updateUserImage(c *gin.Context) {
	claims := getAccessClaims(c)
	if claims == nil {
		return
	}

	var user database.User
	if err := s.db.First(&user, claims.Subject).Error; err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	file, fileHeader, err := c.Request.FormFile("image")
	if err != nil {
		utils.Fail(c, &utils.APIError{
			Code:    http.StatusBadRequest,
			Message: "Profile image is required",
		}, err)
		return
	}
	defer file.Close()

	// Upload to S3
	imageURL, err := s.s3.UploadImage(c, file, fileHeader)
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	err = s.s3.DeleteImageByURL(c, user.Image)
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	user.Image = imageURL
	if err := s.db.Save(&user).Error; err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	utils.Success(c, "Profile image updated successfully", nil)
}

func (s *Server) deleteUser(c *gin.Context) {
	claims := getAccessClaims(c)
	if claims == nil {
		return
	}

	err := s.db.Unscoped().Delete(&database.User{}, claims.Subject).Error
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	utils.Success(c, "user is deleted successfully", nil)
}
