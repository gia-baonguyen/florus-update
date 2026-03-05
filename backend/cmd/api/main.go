package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/florus/backend/internal/config"
	"github.com/florus/backend/internal/router"
	"github.com/florus/backend/migrations"
	"github.com/florus/backend/pkg/database"
)

func main() {
	// Load configuration
	cfg := config.Load()

	log.Printf("Starting Florus Backend API...")
	log.Printf("Server Mode: %s", cfg.Server.Mode)

	// Connect to database
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// CLI mode (one-shot commands). Useful for reseeding inside Docker.
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			if err := migrations.AutoMigrate(db); err != nil {
				log.Fatalf("Failed to run migrations: %v", err)
			}
			log.Println("Migrate completed")
			return
		case "seed":
			if err := migrations.Seed(db); err != nil {
				log.Fatalf("Failed to seed database: %v", err)
			}
			log.Println("Seed completed")
			return
		case "reseed":
			if err := migrations.AutoMigrate(db); err != nil {
				log.Fatalf("Failed to run migrations: %v", err)
			}
			if err := migrations.ForceSeed(db); err != nil {
				log.Fatalf("Failed to reseed database: %v", err)
			}
			log.Println("Reseed completed")
			return
		}
	}

	// Run migrations
	if err := migrations.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed database
	if err := migrations.Seed(db); err != nil {
		log.Printf("Warning: Failed to seed database: %v", err)
	}

	// Setup router
	r := router.Setup(db, cfg)

	// Create server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server listening on http://localhost:%s", cfg.Server.Port)
		log.Printf("API Documentation: http://localhost:%s/api/health", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
