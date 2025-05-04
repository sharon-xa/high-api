package server

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	engine := gin.Default()

	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Add your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true, // Enable cookies/auth
	}))

	s.registerPublicRoutes(engine)
	s.registerUserRoutes(engine)
	s.registerAdminRoutes(engine)

	return engine
}

func (s *Server) registerPublicRoutes(e *gin.Engine) {
	auth := e.Group("/auth")

	auth.POST("/register", s.register)
	auth.POST("/verify-email", s.verifyEmail)
	auth.POST("/login")
	auth.POST("/forgot-password")
	auth.POST("/reset-password")

	posts := e.Group("/posts")
	posts.GET("")
	posts.GET("/:id", s.getPost)
	posts.GET("/:id/comments")

	categories := e.Group("/categories")
	categories.GET("")

	users := e.Group("/users")
	users.GET("/:id")

	tags := e.Group("/tags")
	tags.GET("")
	tags.GET("/:name") // Retrieve all posts associated with a specific tag.
}

func (s *Server) registerUserRoutes(e *gin.Engine) {
	users := e.Group("/users")
	users.PUT("/:id")
	users.DELETE("/:id")

	posts := e.Group("/posts")
	posts.POST("", s.addPost)
	posts.PUT("/:id")
	posts.DELETE("/:id")
	posts.POST("/:id/comments")

	comments := e.Group("/comments")
	comments.PUT("/:id")
	comments.DELETE("/:id")
}

func (s *Server) registerAdminRoutes(e *gin.Engine) {
	admin := e.Group("/admin")
	admin.GET("/users")               // retrieve all users
	admin.POST("/users/{id}/ban")     // ban a user
	admin.POST("/users/{id}/promote") // promote a user to admin

	admin.DELETE("/posts/:id")

	admin.GET("/comments")
	admin.DELETE("/comments/:id")

	admin.POST("/category")
	admin.PUT("/category/:id")
	admin.DELETE("/category/:id")

	admin.PUT("/tags/:id")
	admin.DELETE("/tags/:id")

	admin.GET("/stat")
}
