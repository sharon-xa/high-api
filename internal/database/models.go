package database

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name      string
	Gender    string
	Image     string
	Bio       string
	Email     string `gorm:"unique"`
	Password  string
	Role      string    `gorm:"default:'user'"`
	Verified  bool      `gorm:"default:false"`
	Banned    bool      `gorm:"default:false"`
	Birthdate time.Time `gorm:"type:date"      json:"birthdate"`

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
	ID           uint      `gorm:"primaryKey"`
	UserID       uint      `gorm:"index;not null"`
	User         User      `gorm:"constraint:OnDelete:CASCADE;"`
	RefreshToken string    `gorm:"not null"`
	ExpiresAt    time.Time `gorm:"not null"`
	Revoked      bool      `gorm:"default:false"`
	DeviceID     string    `gorm:"not null;unique"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
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
	ID    uint   `gorm:"primaryKey"           json:"id"`
	Name  string `gorm:"unique"               json:"name"`
	Posts []Post `gorm:"many2many:post_tags;" json:"-"`
}

type PostTag struct {
	PostID uint `gorm:"primaryKey"`
	TagID  uint `gorm:"primaryKey"`
}

type Category struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"unique"     json:"name"`

	Posts []Post `json:"-"`
}
