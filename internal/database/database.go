package database

import (
	"fmt"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type service struct {
	db *gorm.DB
}

var (
	database   = os.Getenv("DB_DATABASE")
	password   = os.Getenv("DB_PASSWORD")
	username   = os.Getenv("DB_USERNAME")
	port       = os.Getenv("DB_PORT")
	host       = os.Getenv("DB_HOST")
	schema     = os.Getenv("DB_SCHEMA")
	tz         = os.Getenv("DB_TIMEZONE")
	dbInstance *service
)

func New() *service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=%s",
		host,
		username,
		password,
		database,
		port,
		tz,
	)
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
