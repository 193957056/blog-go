package config

import (
	"database/sql"
	"log"

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
