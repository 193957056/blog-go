package routes

import (
	"lumina-blog/handlers"

	"github.com/gin-gonic/gin"
)

func PostRoutes(r *gin.RouterGroup) {
	r.GET("/posts", handlers.GetPosts)
	r.GET("/posts/:slug", handlers.GetPost)
	r.POST("/posts", handlers.CreatePost)
	r.PATCH("/posts/:id", handlers.UpdatePost)
	r.DELETE("/posts/:id", handlers.DeletePost)
}

func CategoryRoutes(r *gin.RouterGroup) {
	r.GET("/categories", handlers.GetCategories)
	r.POST("/categories", handlers.CreateCategory)
	r.DELETE("/categories/:id", handlers.DeleteCategory)
}

func TagRoutes(r *gin.RouterGroup) {
	r.GET("/tags", handlers.GetTags)
}

func StatsRoutes(r *gin.RouterGroup) {
	r.GET("/stats", handlers.GetStats)
}
