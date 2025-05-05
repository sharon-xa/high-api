package server

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sharon-xa/high-api/internal/auth"
	"github.com/sharon-xa/high-api/internal/database"
	"github.com/sharon-xa/high-api/internal/utils"
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

type loginReq struct {
	Email    string `json:"email"    binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginRes struct {
	AccessToken string `json:"accessToken"`
	Role        string `json:"role"`
}

func (s *Server) login(c *gin.Context) {
	var req loginReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		utils.Fail(c, utils.ErrBadRequest, err)
		return
	}

	u := database.User{}
	if err := s.db.Where("email = ?", req.Email).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, utils.ErrUnauthorized, nil)
			return
		}
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	err = utils.VerifyPassword(u.Password, req.Password)
	if err != nil {
		utils.Fail(c, utils.ErrUnauthorized, nil)
		return
	}

	if !u.Verified {
		utils.Fail(
			c,
			utils.NewAPIError(http.StatusUnauthorized, "your account isn't verified yet"),
			nil,
		)
		return
	}

	// generate refresh and access tokens
	deviceId := getHeader(c, "Device-ID")
	if deviceId == "" {
		return
	}

	accessToken, err := auth.GenerateAccessToken(
		strconv.Itoa(int(u.ID)),
		u.Role,
		s.env.AccessTokenSecret,
		s.env.AccessTokenExpInMin,
	)
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	refreshToken, err := auth.GenerateRefreshToken(
		strconv.Itoa(int(u.ID)),
		s.env.RefreshTokenSecret,
		s.env.RefreshTokenExpInDays,
	)
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	// get the refresh token by device id
	r := database.RefreshToken{}
	result := s.db.Where("device_id = ?", deviceId).First(&r)

	noPrevRefreshToken := errors.Is(result.Error, gorm.ErrRecordNotFound)
	if result.Error != nil && !noPrevRefreshToken {
		utils.Fail(c, utils.ErrInternal, result.Error)
		return
	}

	// hash the generated refresh token
	hashedRefreshToken, err := utils.HashToken(refreshToken, s.env.RefreshTokenSecret)
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	expiresAt := time.Now().Add((time.Hour * 24) * time.Duration(s.env.RefreshTokenExpInDays))

	// if there's a previous token then update it
	if !noPrevRefreshToken {

		if r.UserID != u.ID {
			utils.Fail(c, utils.ErrBadRequest, errors.New("refresh token might be stolen"))
			return
		}

		if r.DeviceID != deviceId {
			utils.Fail(c, utils.ErrBadRequest, errors.New("refresh token might be stolen"))
			return
		}

		r.RefreshToken = hashedRefreshToken
		r.ExpiresAt = expiresAt
		if err := s.db.Save(&r).Error; err != nil {
			utils.Fail(c, utils.ErrInternal, err)
			return
		}

	} else {
		// new device is being used to login

		r := database.RefreshToken{
			UserID:       u.ID,
			RefreshToken: hashedRefreshToken,
			ExpiresAt:    expiresAt,
			Revoked:      false,
			DeviceID:     deviceId,
		}

		err := s.db.Create(&r).Error
		if !(utils.ValidateFKey(c, err, "user_id")) {
			return
		}

		if !(utils.ValidateUniqueness(c, err, "device_id")) {
			return
		}

		if err != nil {
			utils.Fail(c, utils.ErrInternal, err)
			return
		}

		var sessionCount int64
		err = s.db.Model(&database.RefreshToken{}).Where("user_id = ?", u.ID).Count(&sessionCount).Error
		if err != nil {
			utils.Fail(c, utils.ErrInternal, err)
			return
		}

		if sessionCount > 5 {
			gonnaBeDeletedToken := database.RefreshToken{}

			err := s.db.Where("user_id = ?", u.ID).Order("revoked DESC, created_at ASC").First(&gonnaBeDeletedToken).Error
			if err != nil {
				utils.Fail(c, utils.ErrInternal, err)
				return
			}

			if err := s.db.Delete(&gonnaBeDeletedToken).Error; err != nil {
				utils.Fail(c, utils.ErrInternal, err)
				return
			}
		}
	}

	tokenExp := 60 * 60 * 24 * s.env.RefreshTokenExpInDays
	setCookie(
		c,
		"refresh_token",
		refreshToken,
		tokenExp,
		s.env,
	)

	c.JSON(http.StatusOK, loginRes{
		Role:        u.Role,
		AccessToken: accessToken,
	})
}

func (s *Server) logout(c *gin.Context) {
	deviceId := getHeader(c, "Device-ID")
	if deviceId == "" {
		return
	}

	var refreshToken database.RefreshToken

	err := s.db.Where("device_id = ?", deviceId).First(&refreshToken).Error
	if err != nil {
		// Check if the refresh token doesn't exists
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	refreshToken.Revoked = true
	err = s.db.Save(&refreshToken).Error
	if err != nil {
		utils.Fail(c, utils.ErrInternal, err)
		return
	}

	utils.Success(c, "session revoked successfully", nil)
}
