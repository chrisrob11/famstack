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
	"famstack/internal/jobs"
	"famstack/internal/jobsystem"
	"famstack/internal/server"
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

	// Configure job system
	config := jobsystem.DefaultConfig()
	config.DatabasePath = *dbPath
	config.WorkerConcurrency = map[string]int{
		"default":         3,
		"task_generation": 2,
	}

	// Create job system
	jobSystem := jobsystem.NewSQLiteJobSystem(config, db)

	// Register job handlers
	jobSystem.Register("generate_scheduled_task", jobs.NewTaskGenerationHandler(db))

	// Create and start server
	srv := server.New(db, jobSystem, &server.Config{
		Port: *port,
		Dev:  *dev,
	})

	// Start job system
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Println("Starting job system...")
		if err := jobSystem.Start(ctx); err != nil {
			log.Fatalf("Job system failed to start: %v", err)
		}
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
	log.Println("Shutting down server and job system...")

	// Stop job system
	jobSystem.Stop()
	log.Println("Job system stopped")

	// Stop server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	} else {
		log.Println("Server stopped gracefully")
	}
}
