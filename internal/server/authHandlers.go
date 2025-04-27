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

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&u).Error; err != nil {
			if !utils.ValidateUniqueness(c, err, "user") {
				return err
			}
			utils.Fail(c, utils.ErrInternal, tx.Error)
			return err
		}

		otp := utils.GenerateRandomOTP()
		expTime := time.Now().Add(time.Minute * time.Duration(s.env.OtpExpMin))
		a := database.AccountVerificationOTP{
			UserID:    u.ID,
			OTP:       otp,
			ExpiresAt: expTime,
		}

		if err := tx.Create(&a).Error; err != nil {
			utils.Fail(c, utils.ErrInternal, err)
			return err
		}

		err = auth.SendVerificationEmail(u.Email, otp, s.env)
		if err != nil {
			utils.Fail(c, utils.ErrInternal, err)
			return err
		}
		return nil
	})
	if err != nil {
		return
	}

	utils.Success(c, "account created", nil)
}

type verifyEmailReq struct {
	OTP   string `json:"otp"   binding:"required"`
	Email string `json:"email" binding:"required"`
}

func (s *Server) verifyEmail(c *gin.Context) {
	var req verifyEmailReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		utils.Fail(c, utils.ErrBadRequest, err)
		return
	}

	u := database.User{}
	if err = s.db.Where("email = ?", req.Email).First(&u).Error; err != nil {
		utils.Fail(c, utils.ErrUnauthorized, err)
		return
	}

	if u.Verified {
		utils.Success(c, "your account has already been verified", nil)
		return
	}

	otp := database.AccountVerificationOTP{}
	if err := s.db.Where("user_id = ?", u.ID).First(&otp).Error; err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	if time.Now().After(otp.ExpiresAt) {
		utils.Fail(c, utils.NewAPIError(http.StatusUnauthorized, "OTP is expired"), nil)
		return
	}

	if otp.OTP != req.OTP {
		utils.Fail(
			c,
			&utils.APIError{Code: http.StatusBadRequest, Message: "wrong OTP, please try again"},
			nil,
		)
		return
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		u.Verified = true
		if err := tx.Save(&u).Error; err != nil {
			utils.Fail(c, utils.ErrInternal, err)
			return err
		}
		if err := tx.Delete(&otp).Error; err != nil {
			utils.Fail(c, utils.ErrInternal, err)
			return err
		}
		return nil
	})
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	utils.Success(c, "account verified successfully", nil)
}
