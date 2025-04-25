package server

import (
	"fmt"
	"net/http"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/refine-software/high-api/internal/config"
	"github.com/refine-software/high-api/internal/database"
	"gorm.io/gorm"
)

type Server struct {
	port int

	db  *gorm.DB
	env *config.Env
}

func NewServer() *http.Server {
	env := config.NewEnv()

	dbService := database.New(env.DSN)
	dbService.AutoMigrate()

	NewServer := &Server{
		port: env.Port,
		db:   dbService.DB(),
		env:  env,
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
