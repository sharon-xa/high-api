package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/sharon-xa/high-api/internal/config"
	"github.com/sharon-xa/high-api/internal/database"
	"github.com/sharon-xa/high-api/internal/s3"
	"gorm.io/gorm"
)

type Server struct {
	port int

	db  *gorm.DB
	env *config.Env
	s3  *s3.S3Storage
}

func NewServer() *http.Server {
	env := config.NewEnv()

	dbService := database.New(env.DSN)
	dbService.AutoMigrate()

	s3Storage, err := s3.NewS3Storage(env)
	if err != nil {
		log.Println("Couldn't initialize s3")
		log.Fatalln(err)
	}

	NewServer := &Server{
		port: env.Port,
		db:   dbService.DB(),
		env:  env,
		s3:   s3Storage,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
