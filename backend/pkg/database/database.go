package database

import (
	"fmt"
	"log"

	"github.com/florus/backend/internal/config"
	oracle "github.com/dzwvip/gorm-oracle"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	var err error

	// Oracle connection string format for go-ora:
	// oracle://user:password@host:port/service_name
	dsn := fmt.Sprintf("oracle://%s:%s@%s:%s/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Service,
	)

	DB, err = gorm.Open(oracle.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Printf("Failed to connect to Oracle database: %v", err)
		return nil, err
	}

	// Set connection pool
	sqlDB, _ := DB.DB()
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)

	log.Println("Oracle database connected successfully")
	return DB, nil
}

func GetDB() *gorm.DB {
	return DB
}
