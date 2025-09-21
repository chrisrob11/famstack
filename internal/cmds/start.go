package cmds

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"

	"famstack/internal/auth"
	"famstack/internal/calendar"
	"famstack/internal/config"
	"famstack/internal/database"
	"famstack/internal/encryption"
	"famstack/internal/jobs"
	"famstack/internal/jobsystem"
	"famstack/internal/oauth"
	"famstack/internal/server"
	"famstack/internal/services"
)

// StartCommand returns the start command configuration
func StartCommand() *cli.Command {
	return &cli.Command{
		Name:    "start",
		Aliases: []string{"s"},
		Usage:   "Start the FamStack server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Value: "config.json",
				Usage: "Configuration file path",
			},
			&cli.StringFlag{
				Name:  "port",
				Value: "8080",
				Usage: "Server port",
			},
			&cli.StringFlag{
				Name:  "db",
				Value: "famstack.db",
				Usage: "Database file path",
			},
			&cli.BoolFlag{
				Name:  "dev",
				Usage: "Enable development mode",
			},
			&cli.BoolFlag{
				Name:  "migrate-up",
				Usage: "Run database migrations up",
			},
			&cli.BoolFlag{
				Name:  "migrate-down",
				Usage: "Run database migrations down",
			},
		},
		Action: startServer,
	}
}

// startServer handles the start command implementation
func startServer(ctx *cli.Context) error {
	port := ctx.String("port")
	dbPath := ctx.String("db")
	migrateUp := ctx.Bool("migrate-up")
	migrateDown := ctx.Bool("migrate-down")
	dev := ctx.Bool("dev")

	// Initialize database
	db, err := database.New(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Handle migration commands
	if migrateUp {
		if migErr := db.MigrateUp(); migErr != nil {
			return fmt.Errorf("failed to run migrations up: %w", migErr)
		}
		log.Println("Migrations completed successfully")
		return nil
	}

	if migrateDown {
		if migErr := db.MigrateDown(); migErr != nil {
			return fmt.Errorf("failed to run migrations down: %w", migErr)
		}
		log.Println("Migrations rolled back successfully")
		return nil
	}

	// Run migrations automatically if not explicitly handling them
	if migErr := db.MigrateUp(); migErr != nil {
		return fmt.Errorf("failed to run automatic migrations: %w", migErr)
	}

	// Initialize configuration manager
	configManager, err := config.NewManager("famstack-config.json")
	if err != nil {
		return fmt.Errorf("failed to initialize config manager: %w", err)
	}
	log.Println("📋 Configuration manager initialized successfully")

	// Initialize encryption service with default configuration
	// For now, use keyring with auto-creation
	encryptionConfig := config.DefaultEncryptionSettings()
	encryptionService, err := encryption.NewService(*encryptionConfig)
	if err != nil {
		// Try to provide helpful error message
		log.Printf("Failed to initialize encryption service: %v", err)
		log.Println("💡 Try running 'famstack encryption generate-key' to create a development key")
		return fmt.Errorf("encryption service initialization failed: %w", err)
	}

	log.Println("🔐 Encryption service initialized successfully")

	// Initialize authentication service using encryption service for JWT signing
	authService := auth.NewService(db, encryptionService, "famstack")
	log.Println("🔑 Authentication service initialized successfully")

	// Initialize service registry with all services
	serviceRegistry := services.NewRegistry(db, encryptionService)
	log.Println("🔧 Service registry initialized successfully")

	// Configure job system
	jobConfig := jobsystem.DefaultConfig()
	jobConfig.DatabasePath = dbPath
	jobConfig.WorkerConcurrency = map[string]int{
		"default":         3,
		"task_generation": 2,
	}

	// Create job system
	jobSystem := jobsystem.NewDBJobSystem(jobConfig, serviceRegistry.Jobs)

	// Initialize OAuth and calendar services for job handlers
	// Get OAuth configuration from config manager
	googleConfig, err := configManager.GetOAuthProvider("google")
	if err != nil {
		log.Printf("Warning: Google OAuth not configured: %v", err)
		googleConfig = nil
	}

	var oauthConfig *oauth.OAuthConfig
	if googleConfig != nil && googleConfig.Configured {
		oauthConfig = &oauth.OAuthConfig{
			Google: &oauth.GoogleConfig{
				ClientID:     googleConfig.ClientID,
				ClientSecret: googleConfig.ClientSecret,
				RedirectURL:  googleConfig.RedirectURL,
				Scopes:       googleConfig.Scopes,
			},
		}
		log.Println("🔗 Google OAuth configured successfully")
	} else {
		log.Println("⚠️  Google OAuth not configured - calendar integration will be unavailable")
		oauthConfig = &oauth.OAuthConfig{} // Empty config
	}
	oauthService := oauth.NewService(serviceRegistry.OAuth, oauthConfig, encryptionService)
	googleClient := calendar.NewGoogleClient(oauthService)

	// Register job handlers
	jobSystem.Register("monthly_task_generation", jobs.NewMonthlyTaskGenerationHandler(serviceRegistry))
	jobSystem.Register("schedule_maintenance", jobs.NewScheduleMaintenanceHandler(serviceRegistry, jobSystem))
	jobSystem.Register("delete_schedule", jobs.NewScheduleDeletionHandler(serviceRegistry))
	calendarSyncHandler := jobs.NewCalendarSyncHandler(serviceRegistry, oauthService, googleClient)
	jobSystem.Register("calendar_sync", calendarSyncHandler.Handle)

	// Create and start server
	srv := server.New(serviceRegistry, jobSystem, authService, configManager, &server.Config{
		Port: port,
		Dev:  dev,
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
	jobCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Println("Starting job system...")
		if err := jobSystem.Start(jobCtx); err != nil {
			log.Fatalf("Job system failed to start: %v", err)
		}
	}()

	// Run immediate startup check for schedules needing generation
	go func() {
		// Wait a moment for job system to be fully started
		time.Sleep(2 * time.Second)

		log.Println("Running startup check for schedules needing task generation...")
		startupKey := "startup_maintenance"
		_, err := jobSystem.Enqueue(&jobsystem.EnqueueRequest{
			QueueName:      "task_generation",
			JobType:        "schedule_maintenance",
			Payload:        map[string]interface{}{},
			Priority:       2, // Higher priority than regular maintenance
			MaxRetries:     3,
			IdempotencyKey: &startupKey, // Prevent multiple startup jobs
		})
		if err != nil {
			log.Printf("Failed to enqueue startup maintenance job: %v", err)
		} else {
			log.Println("Enqueued startup maintenance check")
		}
	}()

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", port)
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

	return nil
}
