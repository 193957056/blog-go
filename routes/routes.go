package routes

import (
	"lumina-blog/handlers"
	"lumina-blog/middleware"

	"github.com/gin-gonic/gin"
)

func PostRoutes(r *gin.RouterGroup) {
	// Public routes
	r.GET("/posts", handlers.GetPosts)
	r.GET("/posts/slug/:slug", handlers.GetPost) // Get by slug

	// Protected routes - require authentication
	admin := r.Group("")
	admin.Use(middleware.AuthMiddleware())
	{
		admin.POST("/posts", handlers.CreatePost)
		admin.GET("/posts/:id", handlers.GetPostByID)
		admin.PATCH("/posts/:id", handlers.UpdatePost)
		admin.DELETE("/posts/:id", handlers.DeletePost)
		admin.POST("/posts/:id/translate", handlers.TranslatePost)
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

// SEORoutes handles SEO-related endpoints
func SEORoutes(r *gin.RouterGroup) {
	r.GET("/seo/analyze", handlers.SEOAnalyze)
}

func AuthRoutes(r *gin.RouterGroup) {
	r.POST("/auth/register", handlers.Register)
	r.POST("/auth/login", handlers.Login)
	r.GET("/auth/me", middleware.AuthMiddleware(), handlers.GetCurrentUser)
}

func AIRoutes(r *gin.RouterGroup) {
	// Initialize AI service
	handlers.InitAIService()

	// AI assistant endpoints
	r.POST("/ai/polish", handlers.AIPolish)
	r.POST("/ai/summary", handlers.AISummary)
	r.POST("/ai/seo", handlers.AISEOSuggestion)
	r.POST("/ai/translate", handlers.AITranslate)
}

// PerformanceRoutes handles performance monitoring endpoints
func PerformanceRoutes(r *gin.RouterGroup) {
	r.GET("/stats/performance", handlers.GetPerformanceStats)
	r.GET("/cache/clear", handlers.ClearCache)
	r.POST("/image/process", handlers.ProcessImage)
	r.GET("/health", handlers.HealthCheck)
}

// BackupRoutes handles backup and restore endpoints
func BackupRoutes(r *gin.RouterGroup) {
	admin := r.Group("")
	admin.Use(middleware.AuthMiddleware())
	{
		admin.GET("/backups", handlers.ListBackups)
		admin.GET("/backups/:id", handlers.GetBackup)
		admin.POST("/backups/:id/restore", handlers.RestoreBackup)
		admin.DELETE("/backups/:id", handlers.DeleteBackup)
	}
}
