package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"famstack/internal/database"
	"famstack/internal/server"
	"famstack/internal/workers"
)

func main() {
	var (
		port        = flag.String("port", "8080", "Server port")
		dbPath      = flag.String("db", "famstack.db", "Database file path")
		migrateUp   = flag.Bool("migrate-up", false, "Run database migrations up")
		migrateDown = flag.Bool("migrate-down", false, "Run database migrations down")
		dev         = flag.Bool("dev", false, "Enable development mode")
	)
	flag.Parse()

	// Initialize database
	db, err := database.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Handle migration commands
	if *migrateUp {
		if err := database.MigrateUp(db); err != nil {
			log.Fatalf("Failed to run migrations up: %v", err)
		}
		log.Println("Migrations completed successfully")
		return
	}

	if *migrateDown {
		if err := database.MigrateDown(db); err != nil {
			log.Fatalf("Failed to run migrations down: %v", err)
		}
		log.Println("Migrations rolled back successfully")
		return
	}

	// Run migrations automatically if not explicitly handling them
	if err := database.MigrateUp(db); err != nil {
		log.Fatalf("Failed to run automatic migrations: %v", err)
	}

	// Create and start server
	srv := server.New(db, &server.Config{
		Port: *port,
		Dev:  *dev,
	})

	// Start queue worker in a goroutine
	queueWorker := workers.NewQueueWorker(db, 5*time.Minute)
	go func() {
		log.Println("Starting queue worker...")
		queueWorker.Start()
	}()

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", *port)
		if err := srv.Start(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// Graceful shutdown
	log.Println("Shutting down server and queue worker...")

	// Stop queue worker
	queueWorker.Stop()
	log.Println("Queue worker stopped")

	// Stop server
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	} else {
		log.Println("Server stopped gracefully")
	}
}
