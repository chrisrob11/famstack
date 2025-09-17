package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"famstack/internal/auth"
	"famstack/internal/database"
	"famstack/internal/handlers"
	"famstack/internal/handlers/api"
	"famstack/internal/jobsystem"
	"famstack/internal/services"
)

// Config holds server configuration
type Config struct {
	Port string
	Dev  bool
}

// Server represents the HTTP server
type Server struct {
	db          *database.DB
	jobSystem   *jobsystem.SQLiteJobSystem
	authService *auth.Service
	config      *Config
	server      *http.Server
}

// New creates a new server instance
func New(db *database.DB, jobSystem *jobsystem.SQLiteJobSystem, authService *auth.Service, config *Config) *Server {
	s := &Server{
		db:          db,
		jobSystem:   jobSystem,
		authService: authService,
		config:      config,
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
	// Initialize services
	familyMemberService := services.NewFamilyMemberService(s.db.DB)

	// Initialize handlers
	pageHandler := handlers.NewPageHandler(s.db, s.authService)
	taskAPIHandler := api.NewTaskAPIHandler(s.db)
	familyAPIHandler := api.NewFamilyAPIHandler(s.db)
	familyMemberAPIHandler := api.NewFamilyMemberAPIHandler(familyMemberService)
	scheduleAPIHandler := api.NewScheduleHandlerWithJobSystem(s.db, s.jobSystem)
	calendarAPIHandler := api.NewCalendarAPIHandler(s.db)
	authHandler := auth.NewHandlers(s.authService)
	authMiddleware := auth.NewMiddleware(s.authService)

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

	// Page routes
	mux.HandleFunc("/login", pageHandler.ServePage) // Login page - no auth required

	// Protected app pages - require authentication
	mux.Handle("/tasks", authMiddleware.RequireAuth(http.HandlerFunc(pageHandler.ServePage)))
	mux.Handle("/daily", authMiddleware.RequireAuth(http.HandlerFunc(pageHandler.ServePage)))
	mux.Handle("/family/setup", authMiddleware.RequireAuth(http.HandlerFunc(pageHandler.ServePage)))
	mux.Handle("/family", authMiddleware.RequireAuth(http.HandlerFunc(pageHandler.ServePage)))
	mux.Handle("/schedules", authMiddleware.RequireAuth(http.HandlerFunc(pageHandler.ServePage)))

	// JSON API routes - protected with authentication
	mux.Handle("/api/v1/tasks", authMiddleware.RequireEntityAction(auth.EntityTask, auth.ActionRead)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "GET":
				taskAPIHandler.ListTasks(w, r)
			case "POST":
				authMiddleware.RequireEntityAction(auth.EntityTask, auth.ActionCreate)(
					http.HandlerFunc(taskAPIHandler.CreateTask)).ServeHTTP(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})))

	mux.Handle("/api/v1/tasks/", authMiddleware.RequireAuth(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "PATCH":
				authMiddleware.RequireEntityAction(auth.EntityTask, auth.ActionUpdate)(
					http.HandlerFunc(taskAPIHandler.UpdateTask)).ServeHTTP(w, r)
			case "DELETE":
				authMiddleware.RequireEntityAction(auth.EntityTask, auth.ActionDelete)(
					http.HandlerFunc(taskAPIHandler.DeleteTask)).ServeHTTP(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})))

	// Family API routes - protected with authentication
	mux.Handle("/api/v1/families", authMiddleware.RequireEntityAction(auth.EntityFamily, auth.ActionRead)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "GET":
				familyAPIHandler.ListFamilies(w, r)
			case "POST":
				authMiddleware.RequireEntityAction(auth.EntityFamily, auth.ActionCreate)(
					http.HandlerFunc(familyAPIHandler.CreateFamily)).ServeHTTP(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})))

	mux.Handle("/api/v1/users", authMiddleware.RequireEntityAction(auth.EntityUser, auth.ActionRead)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "GET":
				familyAPIHandler.ListUsers(w, r)
			case "POST":
				authMiddleware.RequireEntityAction(auth.EntityUser, auth.ActionCreate)(
					http.HandlerFunc(familyAPIHandler.CreateUser)).ServeHTTP(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})))

	mux.Handle("/api/v1/users/", authMiddleware.RequireAuth(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "GET":
				familyAPIHandler.GetUser(w, r)
			case "PATCH":
				authMiddleware.RequireEntityAction(auth.EntityUser, auth.ActionUpdate)(
					http.HandlerFunc(familyAPIHandler.UpdateUser)).ServeHTTP(w, r)
			case "DELETE":
				authMiddleware.RequireEntityAction(auth.EntityUser, auth.ActionDelete)(
					http.HandlerFunc(familyAPIHandler.DeleteUser)).ServeHTTP(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})))

	// Family Member API routes - protected with authentication
	mux.Handle("/api/v1/families/members", authMiddleware.RequireAuth(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "GET":
				familyMemberAPIHandler.GetFamilyMembersWithStats(w, r)
			case "POST":
				authMiddleware.RequireEntityAction(auth.EntityFamily, auth.ActionUpdate)(
					http.HandlerFunc(familyMemberAPIHandler.CreateFamilyMember)).ServeHTTP(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})))

	mux.Handle("/api/v1/families/members/", authMiddleware.RequireAuth(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "GET":
				familyMemberAPIHandler.GetFamilyMember(w, r)
			case "PATCH":
				authMiddleware.RequireEntityAction(auth.EntityFamily, auth.ActionUpdate)(
					http.HandlerFunc(familyMemberAPIHandler.UpdateFamilyMember)).ServeHTTP(w, r)
			case "DELETE":
				authMiddleware.RequireEntityAction(auth.EntityFamily, auth.ActionUpdate)(
					http.HandlerFunc(familyMemberAPIHandler.DeleteFamilyMember)).ServeHTTP(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})))

	// Schedule API routes - protected with authentication
	mux.Handle("/api/v1/schedules", authMiddleware.RequireEntityAction(auth.EntitySchedule, auth.ActionRead)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "GET":
				scheduleAPIHandler.ListSchedules(w, r)
			case "POST":
				authMiddleware.RequireEntityAction(auth.EntitySchedule, auth.ActionCreate)(
					http.HandlerFunc(scheduleAPIHandler.CreateSchedule)).ServeHTTP(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})))

	mux.Handle("/api/v1/schedules/", authMiddleware.RequireAuth(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "GET":
				scheduleAPIHandler.GetSchedule(w, r)
			case "PATCH":
				authMiddleware.RequireEntityAction(auth.EntitySchedule, auth.ActionUpdate)(
					http.HandlerFunc(scheduleAPIHandler.UpdateSchedule)).ServeHTTP(w, r)
			case "DELETE":
				authMiddleware.RequireEntityAction(auth.EntitySchedule, auth.ActionDelete)(
					http.HandlerFunc(scheduleAPIHandler.DeleteSchedule)).ServeHTTP(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})))

	// Calendar API routes - protected with authentication
	mux.Handle("/api/v1/calendar/events", authMiddleware.RequireEntityAction(auth.EntityCalendar, auth.ActionRead)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "GET":
				calendarAPIHandler.GetEvents(w, r)
			case "POST":
				authMiddleware.RequireEntityAction(auth.EntityCalendar, auth.ActionCreate)(
					http.HandlerFunc(calendarAPIHandler.CreateEvent)).ServeHTTP(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})))

	mux.Handle("/api/v1/calendar/events/", authMiddleware.RequireAuth(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "GET":
				calendarAPIHandler.GetEvent(w, r)
			case "PATCH":
				authMiddleware.RequireEntityAction(auth.EntityCalendar, auth.ActionUpdate)(
					http.HandlerFunc(calendarAPIHandler.UpdateEvent)).ServeHTTP(w, r)
			case "DELETE":
				authMiddleware.RequireEntityAction(auth.EntityCalendar, auth.ActionDelete)(
					http.HandlerFunc(calendarAPIHandler.DeleteEvent)).ServeHTTP(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})))

	// Authentication API routes
	mux.HandleFunc("/auth/login", authHandler.HandleLogin)
	mux.HandleFunc("/auth/logout", authHandler.HandleLogout)
	mux.HandleFunc("/auth/downgrade", authHandler.HandleDowngrade)
	mux.HandleFunc("/auth/upgrade", authHandler.HandleUpgrade)
	mux.HandleFunc("/auth/refresh", authHandler.HandleRefresh)
	mux.HandleFunc("/auth/me", authHandler.HandleMe)

	// Root route serves daily page - requires authentication
	mux.Handle("/", authMiddleware.RequireAuth(http.HandlerFunc(pageHandler.ServePage)))
}
