package routes

import (
	"lumina-blog/handlers"
	"lumina-blog/middleware"

	"github.com/gin-gonic/gin"
)

func PostRoutes(r *gin.RouterGroup) {
	// Public routes
	r.GET("/posts", handlers.GetPosts)
	r.GET("/posts/:slug", handlers.GetPost)

	// Protected routes - require authentication
	admin := r.Group("")
	admin.Use(middleware.AuthMiddleware())
	{
		admin.POST("/posts", handlers.CreatePost)
		admin.PATCH("/posts/:id", handlers.UpdatePost)
		admin.DELETE("/posts/:id", handlers.DeletePost)
	}
}

func CategoryRoutes(r *gin.RouterGroup) {
	// Public routes
	r.GET("/categories", handlers.GetCategories)

	// Protected routes - require authentication
	admin := r.Group("")
	admin.Use(middleware.AuthMiddleware())
	{
		admin.POST("/categories", handlers.CreateCategory)
		admin.DELETE("/categories/:id", handlers.DeleteCategory)
	}
}

func TagRoutes(r *gin.RouterGroup) {
	r.GET("/tags", handlers.GetTags)
}

func StatsRoutes(r *gin.RouterGroup) {
	r.GET("/stats", handlers.GetStats)
}

func AuthRoutes(r *gin.RouterGroup) {
	r.POST("/auth/register", handlers.Register)
	r.POST("/auth/login", handlers.Login)
	r.GET("/auth/me", middleware.AuthMiddleware(), handlers.GetCurrentUser)
}
