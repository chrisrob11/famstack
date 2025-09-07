package handlers

import (
	"database/sql"
	"html"
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"famstack/internal/database"
	"famstack/internal/models"
	"famstack/internal/validation"
)

type TaskHandler struct {
	db *database.DB
}

func NewTaskHandler(db *database.DB) *TaskHandler {
	return &TaskHandler{db: db}
}

type TaskListData struct {
	Tasks       []TaskWithUser
	TasksByUser map[string]UserColumn
	Date        string
}

type UserColumn struct {
	User  models.User
	Tasks []TaskWithUser
}

type TaskWithUser struct {
	models.Task
	AssignedToName string
	CreatedByName  string
}

func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	// For now, show tasks for the first family (we'll add auth later)
	familyID := "fam1"

	// Get today's date for filtering daily tasks
	today := time.Now().Format("2006-01-02")

	query := `
		SELECT 
			t.id, t.family_id, t.assigned_to, t.title, t.description, t.task_type,
			t.status, t.priority, t.due_date, t.points, t.created_by, t.created_at, t.completed_at,
			COALESCE(assigned_user.name, '') as assigned_to_name,
			creator.name as created_by_name
		FROM tasks t
		LEFT JOIN users assigned_user ON t.assigned_to = assigned_user.id
		JOIN users creator ON t.created_by = creator.id
		WHERE t.family_id = ? 
		AND (t.due_date IS NULL OR DATE(t.due_date) <= ? OR t.status = 'pending')
		AND t.status != 'completed'
		ORDER BY t.status ASC, t.priority DESC, t.created_at DESC
	`

	rows, err := h.db.Query(query, familyID, today)
	if err != nil {
		http.Error(w, "Failed to fetch tasks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []TaskWithUser
	for rows.Next() {
		var task TaskWithUser
		var dueDate, completedAt sql.NullTime

		scanErr := rows.Scan(
			&task.ID, &task.FamilyID, &task.AssignedTo, &task.Title, &task.Description,
			&task.TaskType, &task.Status, &task.Priority, &dueDate, &task.Points,
			&task.CreatedBy, &task.CreatedAt, &completedAt,
			&task.AssignedToName, &task.CreatedByName,
		)
		if scanErr != nil {
			http.Error(w, "Failed to scan task", http.StatusInternalServerError)
			return
		}

		if dueDate.Valid {
			task.DueDate = &dueDate.Time
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}

		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, "Error reading tasks", http.StatusInternalServerError)
		return
	}

	// Get all family members
	usersQuery := "SELECT id, family_id, name, email, role, created_at FROM users WHERE family_id = ? ORDER BY role DESC, name ASC"
	userRows, err := h.db.Query(usersQuery, familyID)
	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}
	defer userRows.Close()

	var users []models.User
	for userRows.Next() {
		var user models.User
		if scanErr := userRows.Scan(&user.ID, &user.FamilyID, &user.Name, &user.Email, &user.Role, &user.CreatedAt); scanErr != nil {
			http.Error(w, "Failed to scan user", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	// Group tasks by user
	tasksByUser := make(map[string]UserColumn)

	// Initialize columns for each user (including those with no tasks)
	for _, user := range users {
		tasksByUser[user.ID] = UserColumn{
			User:  user,
			Tasks: []TaskWithUser{},
		}
	}

	// Add tasks to appropriate user columns
	for _, task := range tasks {
		if task.AssignedTo != nil {
			if column, exists := tasksByUser[*task.AssignedTo]; exists {
				column.Tasks = append(column.Tasks, task)
				tasksByUser[*task.AssignedTo] = column
			}
		} else {
			// Handle unassigned tasks - create a special "Unassigned" column
			if _, exists := tasksByUser["unassigned"]; !exists {
				tasksByUser["unassigned"] = UserColumn{
					User:  models.User{ID: "unassigned", Name: "Unassigned", Role: "unassigned"},
					Tasks: []TaskWithUser{},
				}
			}
			column := tasksByUser["unassigned"]
			column.Tasks = append(column.Tasks, task)
			tasksByUser["unassigned"] = column
		}
	}

	// Render template
	data := TaskListData{
		Tasks:       tasks,
		TasksByUser: tasksByUser,
		Date:        time.Now().Format("Monday, January 2, 2006"),
	}

	// Load template from file
	tmplPath := filepath.Join("web", "templates", "tasks.html.tmpl")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// NewTaskForm returns an HTMX form for creating a new task
func (h *TaskHandler) NewTaskForm(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	// Get user name for validation
	var userName string
	if userID == "unassigned" {
		userName = "Unassigned"
	} else {
		row := h.db.QueryRow("SELECT name FROM users WHERE id = ?", userID)
		if err := row.Scan(&userName); err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
	}

	// Load template
	tmplPath := filepath.Join("web", "templates", "task-form.html.tmpl")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	// Template data
	data := struct {
		UserID string
	}{
		UserID: userID,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// CreateTask handles POST /api/tasks
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Validate and sanitize title
	title := validation.SanitizeTitle(r.FormValue("title"))
	if err := validation.ValidateTitle(title); err != nil {
		http.Error(w, "Invalid title: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate assigned user
	assignedTo := r.FormValue("assigned_to")
	if err := validation.ValidateUserID(assignedTo); err != nil {
		http.Error(w, "Invalid user ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Verify user exists in database (unless unassigned)
	if assignedTo != "" && assignedTo != "unassigned" {
		var exists bool
		err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ? AND family_id = ?)", assignedTo, "fam1").Scan(&exists)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		if !exists {
			http.Error(w, "Assigned user not found", http.StatusBadRequest)
			return
		}
	}

	if assignedTo == "unassigned" {
		assignedTo = ""
	}

	// Create the task
	familyID := "fam1"   // For now, hardcoded
	createdBy := "user1" // For now, hardcoded (would come from session)

	query := `
		INSERT INTO tasks (family_id, assigned_to, title, task_type, status, priority, created_by, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	var assignedToPtr *string
	if assignedTo != "" {
		assignedToPtr = &assignedTo
	}

	_, err := h.db.Exec(query, familyID, assignedToPtr, title, "todo", "pending", 0, createdBy, time.Now())
	if err != nil {
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	// Return the new task HTML
	safeTitle := html.EscapeString(title)
	taskHTML := `
	<div class="task-item pending todo" draggable="true">
		<div class="task-title">` + safeTitle + `</div>
		<div class="task-meta">
			<span class="task-status status-pending">pending</span>
			<span class="task-type type-todo">todo</span>
		</div>
	</div>`

	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(taskHTML)); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

// CancelTaskForm returns the original Add Task button
func (h *TaskHandler) CancelTaskForm(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	// Load template
	tmplPath := filepath.Join("web", "templates", "add-task-button.html.tmpl")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	// Template data
	data := struct {
		UserID string
	}{
		UserID: userID,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}
