package main

import (
	"log"
	"os"
	"time"

	"lumina-blog/config"
	"lumina-blog/middleware"
	"lumina-blog/models"
	"lumina-blog/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize database
	config.InitDB()

	// Auto migrate models
	config.DB.AutoMigrate(
		&models.Post{},
		&models.Category{},
		&models.Tag{},
		&models.User{},
	)

	// Seed default data
	models.SeedData()

	// Setup router
	r := gin.Default()

	// Add performance middleware for tracking slow requests
	r.Use(middleware.PerformanceMiddleware())

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000", "*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Root endpoint - API status
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":        "Lumina Blog API",
			"version":     "2.0",
			"status":      "running",
			"endpoints":   []string{"/api/posts", "/api/categories", "/api/tags", "/api/stats", "/api/stats/performance", "/api/cache/clear", "/api/image/process", "/api/health", "/api/auth/login", "/api/ai/polish", "/api/ai/summary", "/api/ai/seo", "/api/ai/translate", "/api/seo/analyze"},
			"frontend":    "http://localhost:5173",
			"admin":       "http://localhost:5173/admin",
		})
	})

	// API routes
	api := r.Group("/api")
	{
		// Auth routes (public)
		routes.AuthRoutes(api)
		
		// Public routes
		routes.PostRoutes(api)
		routes.CategoryRoutes(api)
		routes.TagRoutes(api)
		routes.StatsRoutes(api)

		// SEO routes
		routes.SEORoutes(api)

		// AI assistant routes
		routes.AIRoutes(api)

		// Performance monitoring routes
		routes.PerformanceRoutes(api)

		// Backup and restore routes
		routes.BackupRoutes(api)
	}

	// Serve static files
	r.Static("/static", "./static")

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
