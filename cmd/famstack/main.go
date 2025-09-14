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
		if migErr := database.MigrateUp(db); migErr != nil {
			log.Fatalf("Failed to run migrations up: %v", migErr)
		}
		log.Println("Migrations completed successfully")
		return
	}

	if *migrateDown {
		if migErr := database.MigrateDown(db); migErr != nil {
			log.Fatalf("Failed to run migrations down: %v", migErr)
		}
		log.Println("Migrations rolled back successfully")
		return
	}

	// Run migrations automatically if not explicitly handling them
	if migErr := database.MigrateUp(db); migErr != nil {
		log.Fatalf("Failed to run automatic migrations: %v", migErr)
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
	jobSystem.Register("monthly_task_generation", jobs.NewMonthlyTaskGenerationHandler(db))
	jobSystem.Register("schedule_maintenance", jobs.NewScheduleMaintenanceHandler(db, jobSystem))
	jobSystem.Register("delete_schedule", jobs.NewScheduleDeletionHandler(db))

	// Create and start server
	srv := server.New(db, jobSystem, &server.Config{
		Port: *port,
		Dev:  *dev,
	})

	// Set up daily maintenance job scheduling
	err = jobSystem.Schedule(&jobsystem.ScheduleRequest{
		Name:      "daily_schedule_maintenance",
		QueueName: "task_generation",
		JobType:   "schedule_maintenance",
		Payload:   map[string]interface{}{},
		CronExpr:  "0 0 * * *", // Daily at midnight
		Enabled:   true,
	})
	if err != nil {
		log.Printf("Failed to schedule daily maintenance job: %v", err)
	} else {
		log.Println("Scheduled daily maintenance job")
	}

	// Start job system
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Println("Starting job system...")
		if err := jobSystem.Start(ctx); err != nil {
			log.Fatalf("Job system failed to start: %v", err)
		}
	}()

	// Run immediate startup check for schedules needing generation
	go func() {
		// Wait a moment for job system to be fully started
		time.Sleep(2 * time.Second)

		log.Println("Running startup check for schedules needing task generation...")
		_, err := jobSystem.Enqueue(&jobsystem.EnqueueRequest{
			QueueName:  "task_generation",
			JobType:    "schedule_maintenance",
			Payload:    map[string]interface{}{},
			Priority:   2, // Higher priority than regular maintenance
			MaxRetries: 3,
		})
		if err != nil {
			log.Printf("Failed to enqueue startup maintenance job: %v", err)
		} else {
			log.Println("Enqueued startup maintenance check")
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
