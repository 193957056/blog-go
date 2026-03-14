package main

import (
	"log"
	"os"
	"time"

	"lumina-blog/config"
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
	)

	// Seed default data
	models.SeedData()

	// Setup router
	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000", "*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API routes
	api := r.Group("/api")
	{
		// Public routes
		routes.PostRoutes(api)
		routes.CategoryRoutes(api)
		routes.TagRoutes(api)
		routes.StatsRoutes(api)
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
