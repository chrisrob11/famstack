package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"path/filepath"

	"famstack/internal/database"
	"famstack/internal/models"
)

type TaskHandler struct {
	db *database.DB
}

func NewTaskHandler(db *database.DB) *TaskHandler {
	return &TaskHandler{db: db}
}

type TaskListData struct {
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
		ORDER BY t.status ASC, t.priority DESC, t.created_at DESC
	`

	rows, err := h.db.Query(query, familyID)
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

	// Render template
	data := TaskListData{Tasks: tasks}

	// Load template from file
	tmplPath := filepath.Join("web", "templates", "tasks.html.tmpl")
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}
