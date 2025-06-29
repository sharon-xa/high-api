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

	// login - refresh
	auth.POST("/login", s.login)
	auth.POST("/refresh-tokens", s.refreshTokens)

	// password reset
	auth.POST("/forgot-password", s.forgotPassword)
	auth.POST("/reset-password", s.resetPassword)

	// public posts
	posts := e.Group("/posts")
	posts.GET("")
	posts.GET("/:id", s.getPost)
	posts.GET("/:id/comments", s.getCommentsOfPost)

	// public categories
	categories := e.Group("/categories")
	categories.GET("", s.getCategories)

	// public users
	users := e.Group("/users")
	users.GET("/:id/public", s.getUserPublic)

	// public tags
	tags := e.Group("/tags")
	tags.GET("", s.getAllTags)
}

func (s *Server) registerUserRoutes(e *gin.Engine) {
	protected := e.Group("")
	protected.Use(middleware.User(s.env.AccessTokenSecret))

	auth := protected.Group("/auth")
	auth.POST("/logout", s.logout)
	auth.POST("/logout/all", s.logoutAllSessions)

	users := protected.Group("/users")
	users.GET("/me", s.getUser)
	users.PUT("/me", s.updateUser)
	users.PATCH("/me/image", s.updateUserImage)
	users.DELETE("/me", s.deleteUser)

	posts := protected.Group("/posts")
	posts.POST("", s.addPost)
	posts.PUT("/:id", s.updatePost)
	posts.DELETE("/:id", s.deletePost)
	posts.POST("/:id/comment", s.addComment)

	comments := protected.Group("/comments")
	comments.PUT("/:id", s.updateComment)
	comments.DELETE("/:id", s.deleteComment)
}

func (s *Server) registerAdminRoutes(e *gin.Engine) {
	admin := e.Group("/admin")
	admin.Use(middleware.Admin(s.env.AccessTokenSecret))

	admin.GET("/users", s.getAllUsers)
	admin.POST("/users/:id/ban", s.banUser)
	admin.POST("/users/:id/promote", s.promoteUser)

	admin.DELETE("/posts/:id", s.deletePost)

	admin.GET("/comments")
	admin.DELETE("/comments/:id", s.deleteComment)

	admin.POST("/category", s.addCategory)
	admin.PUT("/category/:id", s.updateCategory)
	admin.DELETE("/category/:id", s.deleteCategory)

	admin.PUT("/tags/:id", s.updateTag)
	admin.DELETE("/tags/:id", s.deleteTag)

	admin.GET("/stat")
}
