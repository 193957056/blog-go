package config

import (
	"database/sql"
	"log"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBStats holds database connection pool statistics
type DBStats struct {
	OpenConnections int
	InUse           int
	Idle            int
	WaitCount       int64
	WaitDuration    int64
}

var DB *gorm.DB

// SEOConfig holds SEO-related configuration
var SEOConfig struct {
	SiteName        string
	SiteURL         string
	Description     string
	BaiduTongjiID   string
	GoogleAnalytics string
}

func InitDB() {
	var err error

	// Use logger.Silent to disable SQL log output
	DB, err = gorm.Open(sqlite.Open("lumina.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connected successfully")
	
	// Load SEO config from environment
	LoadSEOConfig()
}

// LoadSEOConfig loads SEO configuration from environment variables
func LoadSEOConfig() {
	SEOConfig.SiteName = os.Getenv("SITE_NAME")
	if SEOConfig.SiteName == "" {
		SEOConfig.SiteName = "Lumina Blog"
	}
	
	SEOConfig.SiteURL = os.Getenv("SITE_URL")
	if SEOConfig.SiteURL == "" {
		SEOConfig.SiteURL = "https://example.com"
	}
	
	SEOConfig.Description = os.Getenv("SITE_DESCRIPTION")
	if SEOConfig.Description == "" {
		SEOConfig.Description = "A beautiful blog"
	}
	
	SEOConfig.BaiduTongjiID = os.Getenv("BAIDU_TONGJI_ID")
	SEOConfig.GoogleAnalytics = os.Getenv("GOOGLE_ANALYTICS_ID")
}

// GetDBStats returns database connection pool statistics
func GetDBStats() DBStats {
	sqlDB, err := DB.DB()
	if err != nil {
		return DBStats{}
	}
	
	stats := sqlDB.Stats()
	return DBStats{
		OpenConnections: stats.OpenConnections,
		InUse:          stats.InUse,
		Idle:           stats.Idle,
		WaitCount:      stats.WaitCount,
		WaitDuration:   int64(stats.WaitDuration),
	}
}

// Helper function to check if record exists
func Exists(model interface{}, conditions ...interface{}) bool {
	result := DB.First(model, conditions...)
	return result.RowsAffected > 0
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// GetRawDB returns the standard library database for direct access
func GetRawDB() (*sql.DB, error) {
	return DB.DB()
}
