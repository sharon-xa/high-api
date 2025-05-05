package database

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name     string
	Gender   string
	Image    string
	Bio      string
	Email    string `gorm:"unique"`
	Password string
	Role     string `gorm:"default:'user'"`
	Verified bool   `gorm:"default:false"`

	Posts                  []Post
	Comments               []Comment
	AccountVerificationOTP AccountVerificationOTP `gorm:"constraint:OnDelete:CASCADE;"`
	RefreshTokens          []RefreshToken
}

type AccountVerificationOTP struct {
	gorm.Model
	UserID    uint `gorm:"uniqueIndex"`
	OTP       string
	ExpiresAt time.Time
}

type PasswordResetToken struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"index"`
	Token     string `gorm:"unique"`
	ExpiresAt time.Time
}

type RefreshToken struct {
	gorm.Model
	UserID       uint      `gorm:"index;not null"`
	User         User      `gorm:"constraint:OnDelete:CASCADE;"`
	RefreshToken string    `gorm:"not null"`
	ExpiresAt    time.Time `gorm:"not null"`
	Revoked      bool      `gorm:"default:false"`
	DeviceID     string    `gorm:"not null;unique"`
}

type Post struct {
	gorm.Model
	UserID     uint
	CategoryID uint
	Title      string
	Content    string
	Image      string

	User     User `gorm:"constraint:OnDelete:CASCADE;"`
	Category Category
	Tags     []Tag `gorm:"many2many:post_tags;"`
	Comments []Comment
}

type Comment struct {
	gorm.Model
	PostID  uint
	UserID  uint
	Content string

	Post Post
	User User
}

type Tag struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"unique"`
	Posts []Post `gorm:"many2many:post_tags;"`
}

type PostTag struct {
	PostID uint `gorm:"primaryKey"`
	TagID  uint `gorm:"primaryKey"`
}

type Category struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"unique"`

	Posts []Post
}
