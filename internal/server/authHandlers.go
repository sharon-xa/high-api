package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/refine-software/high-api/internal/auth"
	"github.com/refine-software/high-api/internal/database"
	"github.com/refine-software/high-api/internal/utils"
	"gorm.io/gorm"
)

type registerReq struct {
	Name     string `json:"name"     binding:"required"`
	Email    string `json:"email"    binding:"required"`
	Gender   string `json:"gender"   binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (s *Server) register(c *gin.Context) {
	var req registerReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		utils.Fail(c, utils.ErrBadRequest, nil)
		return
	}

	hashedPass, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	role := "user"
	if s.env.AdminEmail == req.Email {
		role = "admin"
	}

	u := database.User{
		Name:     req.Name,
		Email:    req.Email,
		Gender:   req.Gender,
		Password: hashedPass,
		Verified: false,
		Role:     role,
	}

	tx := s.db.Begin()
	defer tx.Rollback()
	if err := tx.Error; err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	err = tx.Create(&u).Error
	if err != nil {
		if !utils.ValidateUniqueness(c, err, "user") {
			return
		}
		utils.Fail(c, utils.ErrInternal, tx.Error)
		return
	}

	otp := utils.GenerateRandomOTP()
	expTime := time.Now().Add(time.Minute * time.Duration(s.env.OtpExpMin))
	a := database.AccountVerificationOTP{
		UserID:    u.ID,
		OTP:       otp,
		ExpiresAt: expTime,
	}

	err = tx.Create(&a).Error
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	err = auth.SendVerificationEmail(u.Email, otp, s.env)
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	err = tx.Commit().Error
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	utils.Success(c, "account created", nil)
}

type verifyEmailReq struct {
	OTP string `json:"otp" binding:"required"`
}

func (s *Server) verifyEmail(c *gin.Context) {
}
