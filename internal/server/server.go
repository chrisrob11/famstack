package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"famstack/internal/database"
	"famstack/internal/handlers"
	"famstack/internal/handlers/api"
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
	pageHandler := handlers.NewPageHandler(s.db)
	taskAPIHandler := api.NewTaskAPIHandler(s.db)
	familyAPIHandler := api.NewFamilyAPIHandler(s.db)
	scheduleAPIHandler := api.NewScheduleHandler(s.db)

	// Static file serving
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","message":"Fam-Stack is running"}`)
	})

	// Debug endpoint to test task data server-side
	mux.HandleFunc("/debug/tasks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		// We'll create a simple HTML page showing the task data
		fmt.Fprintf(w, `
			<!DOCTYPE html>
			<html>
			<head><title>Debug Tasks</title></head>
			<body>
				<h1>Debug: Task Data from Server</h1>
				<p>This endpoint shows task data rendered server-side to verify data availability.</p>
				<div id="debug-info">Loading...</div>
					<script>
						// Make API call and display results
						fetch('/api/v1/tasks')
							.then(response => response.json())
							.then(data => {
								document.getElementById('debug-info').innerHTML = '<pre>' + JSON.stringify(data, null, 2) + '</pre>';
							})
							.catch(error => {
								document.getElementById('debug-info').innerHTML = '<p style="color: red;">Error: ' + error + '</p>';
							});
					</script>
			</body>
			</html>
		`)
	})

	// Page routes - single handler for all pages
	mux.HandleFunc("/tasks", pageHandler.ServePage)
	mux.HandleFunc("/family/setup", pageHandler.ServePage)
	mux.HandleFunc("/family", pageHandler.ServePage)
	mux.HandleFunc("/schedules", pageHandler.ServePage)

	// JSON API routes
	mux.HandleFunc("/api/v1/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			taskAPIHandler.ListTasks(w, r)
		case "POST":
			taskAPIHandler.CreateTask(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/tasks/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PATCH":
			taskAPIHandler.UpdateTask(w, r)
		case "DELETE":
			taskAPIHandler.DeleteTask(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Family API routes
	mux.HandleFunc("/api/v1/families", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			familyAPIHandler.ListFamilies(w, r)
		case "POST":
			familyAPIHandler.CreateFamily(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			familyAPIHandler.ListUsers(w, r)
		case "POST":
			familyAPIHandler.CreateUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/users/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			familyAPIHandler.GetUser(w, r)
		case "PATCH":
			familyAPIHandler.UpdateUser(w, r)
		case "DELETE":
			familyAPIHandler.DeleteUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Schedule API routes
	mux.HandleFunc("/api/v1/schedules", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			scheduleAPIHandler.ListSchedules(w, r)
		case "POST":
			scheduleAPIHandler.CreateSchedule(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/schedules/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			scheduleAPIHandler.GetSchedule(w, r)
		case "PATCH":
			scheduleAPIHandler.UpdateSchedule(w, r)
		case "DELETE":
			scheduleAPIHandler.DeleteSchedule(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Root route serves tasks page
	mux.HandleFunc("/", pageHandler.ServePage)
}
