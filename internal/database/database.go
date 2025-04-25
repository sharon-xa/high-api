package database

import (
	"log"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type service struct {
	db *gorm.DB
}

var dbInstance *service

func New(dsn string) *service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		log.Fatal(err)
	}
	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

func (s *service) DB() *gorm.DB {
	return s.db
}

func (s *service) AutoMigrate() {
	err := s.db.AutoMigrate(
		&User{},
		&AccountVerificationOTP{},
		&PasswordResetToken{},
		&Post{},
		&Comment{},
		&Tag{},
		&PostTag{},
		&Category{},
	)
	if err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}
}
