package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"famstack/internal/database"
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
	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","message":"Fam-Stack is running"}`)
	})

	// Root endpoint - temporary placeholder
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Fam-Stack</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<body>
    <h1>Welcome to Fam-Stack!</h1>
    <p>Family task management system is starting up...</p>
    <p><a href="/health">Health Check</a></p>
</body>
</html>`)
	})
}
