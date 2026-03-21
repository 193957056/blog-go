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
	r.GET("/seo/config", handlers.GetSiteConfig)
	r.GET("/seo/baidu-tongji-id", handlers.GetBaiduTongjiID)
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

// CommentRoutes handles comment endpoints
func CommentRoutes(r *gin.RouterGroup) {
	// Public routes
	r.GET("/posts/:id/comments", handlers.GetCommentsByPost)
	r.POST("/comments", handlers.CreateComment)

	// Protected routes - admin only
	admin := r.Group("")
	admin.Use(middleware.AuthMiddleware())
	{
		admin.GET("/comments", handlers.GetAllComments)
		admin.DELETE("/comments/:id", handlers.DeleteComment)
		admin.POST("/comments/:id/approve", handlers.ApproveComment)
	}
}

// SearchRoutes handles search endpoints
func SearchRoutes(r *gin.RouterGroup) {
	r.GET("/search", handlers.Search)
}

// LikeFavoriteRoutes handles like and favorite endpoints
func LikeFavoriteRoutes(r *gin.RouterGroup) {
	// Public routes
	r.GET("/posts/:id/likes", handlers.GetPostLikes)
	r.GET("/posts/:id/favorites", handlers.GetPostFavorites)

	// Protected routes - require authentication
	protected := r.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.POST("/posts/:id/like", handlers.LikePost)
		protected.POST("/posts/:id/favorite", handlers.FavoritePost)
		protected.GET("/user/likes", handlers.GetUserLikes)
		protected.GET("/user/favorites", handlers.GetUserFavorites)
	}
}

// LinkRoutes handles friend link endpoints
func LinkRoutes(r *gin.RouterGroup) {
	// Public route - get approved links
	r.GET("/links", handlers.GetLinks)
	r.POST("/links/apply", handlers.ApplyLink)

	// Admin routes
	admin := r.Group("")
	admin.Use(middleware.AuthMiddleware())
	{
		admin.GET("/admin/links", handlers.GetAllLinks)
		admin.POST("/admin/links", handlers.CreateLink)
		admin.PATCH("/admin/links/:id", handlers.UpdateLink)
		admin.DELETE("/admin/links/:id", handlers.DeleteLink)
	}
}

// AdminRoutes handles admin panel API endpoints
func AdminRoutes(r *gin.RouterGroup) {
	admin := r.Group("")
	admin.Use(middleware.AuthMiddleware())
	{
		// Statistics
		admin.GET("/admin/stats", handlers.GetStats)

		// Post management
		admin.GET("/admin/posts", handlers.AdminPostList)
		admin.PATCH("/admin/posts/:id", handlers.AdminUpdatePost)
		admin.DELETE("/admin/posts/:id", handlers.AdminDeletePost)

		// Comment management
		admin.GET("/admin/comments", handlers.AdminCommentList)
		admin.POST("/admin/comments/:id/approve", handlers.AdminApproveComment)
		admin.POST("/admin/comments/:id/reject", handlers.AdminRejectComment)
		admin.DELETE("/admin/comments/:id", handlers.AdminDeleteComment)
	}
}
