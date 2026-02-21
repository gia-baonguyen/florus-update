package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port         string
	Mode         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Service  string
	SSLMode  string
}

type JWTConfig struct {
	Secret          string
	ExpirationHours int
	Issuer          string
}

func Load() *Config {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("SERVER_MODE", "debug")
	viper.SetDefault("SERVER_READ_TIMEOUT", 10)
	viper.SetDefault("SERVER_WRITE_TIMEOUT", 10)

	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "1521")
	viper.SetDefault("DB_USER", "florus")
	viper.SetDefault("DB_PASSWORD", "florus123")
	viper.SetDefault("DB_SERVICE", "XEPDB1")
	viper.SetDefault("DB_SSL_MODE", "disable")

	viper.SetDefault("JWT_SECRET", "florus-secret-key-change-in-production")
	viper.SetDefault("JWT_EXPIRATION_HOURS", 24)
	viper.SetDefault("JWT_ISSUER", "florus-api")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Config file not found, using defaults and environment variables: %v", err)
	}

	return &Config{
		Server: ServerConfig{
			Port:         viper.GetString("SERVER_PORT"),
			Mode:         viper.GetString("SERVER_MODE"),
			ReadTimeout:  time.Duration(viper.GetInt("SERVER_READ_TIMEOUT")) * time.Second,
			WriteTimeout: time.Duration(viper.GetInt("SERVER_WRITE_TIMEOUT")) * time.Second,
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			Service:  viper.GetString("DB_SERVICE"),
			SSLMode:  viper.GetString("DB_SSL_MODE"),
		},
		JWT: JWTConfig{
			Secret:          viper.GetString("JWT_SECRET"),
			ExpirationHours: viper.GetInt("JWT_EXPIRATION_HOURS"),
			Issuer:          viper.GetString("JWT_ISSUER"),
		},
	}
}
