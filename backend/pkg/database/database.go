package database

import (
	"log"

	"github.com/florus/backend/internal/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	var err error

	// Using SQLite for development (easy setup)
	// For production with Oracle, use: gorm.io/driver/oracle
	// Added busy_timeout and WAL mode for better concurrency
	DB, err = gorm.Open(sqlite.Open("florus.db?_busy_timeout=5000&_journal_mode=WAL"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return nil, err
	}

	// Configure connection pool for SQLite
	sqlDB, err := DB.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1) // SQLite only supports one writer

	log.Println("Database connected successfully")
	return DB, nil
}

func GetDB() *gorm.DB {
	return DB
}
