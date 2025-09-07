package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"famstack/internal/database"
	"famstack/internal/handlers"
)

// Config holds server configuration
type Config struct {
	Port string
	Dev  bool
}

// Server represents the HTTP server
type Server struct {
	db     *database.DB
	config *Config
	server *http.Server
}

// New creates a new server instance
func New(db *database.DB, config *Config) *Server {
	s := &Server{
		db:     db,
		config: config,
	}

	// Set up routes
	mux := http.NewServeMux()
	s.setupRoutes(mux)

	s.server = &http.Server{
		Addr:         ":" + config.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// setupRoutes configures the HTTP routes
func (s *Server) setupRoutes(mux *http.ServeMux) {
	// Initialize handlers
	taskHandler := handlers.NewTaskHandler(s.db)

	// Static file serving
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","message":"Fam-Stack is running"}`)
	})

	// Task routes
	mux.HandleFunc("/tasks", taskHandler.ListTasks)

	// API routes
	mux.HandleFunc("/api/tasks/new", taskHandler.NewTaskForm)
	mux.HandleFunc("/api/tasks/cancel", taskHandler.CancelTaskForm)
	mux.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			taskHandler.CreateTask(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Redirect root to tasks for now
		http.Redirect(w, r, "/tasks", http.StatusFound)
	})
}
