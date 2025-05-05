package server

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sharon-xa/high-api/internal/middleware"
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

	// create account
	auth.POST("/register", s.register)
	auth.POST("/verify-email", s.verifyEmail)
	auth.POST("/resend-verification-otp", s.resendVerificationEmail)

	// login - logout
	auth.POST("/login", s.login)

	// password reset
	auth.POST("/forgot-password", s.forgotPassword)
	auth.POST("/reset-password")

	// public posts
	posts := e.Group("/posts")
	posts.GET("")
	posts.GET("/:id", s.getPost)
	posts.GET("/:id/comments")

	// public categories
	categories := e.Group("/categories")
	categories.GET("")

	// public users
	users := e.Group("/users")
	users.GET("/:id")

	// public tags
	tags := e.Group("/tags")
	tags.GET("")
	tags.GET("/:name") // Retrieve all posts associated with a specific tag.
}

func (s *Server) registerUserRoutes(e *gin.Engine) {
	// logout
	protected := e.Group("")
	protected.Use(middleware.User(s.env.AccessTokenSecret))

	auth := protected.Group("/auth")
	auth.POST("/logout", s.logout)
	auth.POST("/logout/all")

	users := protected.Group("/users")
	users.PUT("/:id")
	users.DELETE("/:id")

	posts := protected.Group("/posts")
	posts.POST("", s.addPost)
	posts.PUT("/:id")
	posts.DELETE("/:id")
	posts.POST("/:id/comments")

	comments := protected.Group("/comments")
	comments.PUT("/:id")
	comments.DELETE("/:id")
}

func (s *Server) registerAdminRoutes(e *gin.Engine) {
	admin := e.Group("/admin")
	admin.Use(middleware.Admin(s.env.AccessTokenSecret))

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
