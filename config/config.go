package config

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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

// Helper function to check if record exists
func Exists(model interface{}, conditions ...interface{}) bool {
	result := DB.First(model, conditions...)
	return result.RowsAffected > 0
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
