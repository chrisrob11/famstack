package workers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"famstack/internal/database"
)

type QueueWorker struct {
	db       *database.DB
	interval time.Duration
	stopCh   chan struct{}
}

type QueuePayload struct {
	ScheduleID string `json:"schedule_id"`
	TargetDate string `json:"target_date"`
}

type TaskSchedule struct {
	ID          string
	FamilyID    string
	CreatedBy   string
	Title       string
	Description string
	TaskType    string
	AssignedTo  *string
	DaysOfWeek  []string
	TimeOfDay   *string
	Priority    int
	Points      int
}

func NewQueueWorker(db *database.DB, interval time.Duration) *QueueWorker {
	return &QueueWorker{
		db:       db,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (qw *QueueWorker) Start() {
	log.Printf("Queue worker started with interval: %v", qw.interval)

	ticker := time.NewTicker(qw.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			qw.processQueue()
		case <-qw.stopCh:
			log.Println("Queue worker stopped")
			return
		}
	}
}

func (qw *QueueWorker) Stop() {
	close(qw.stopCh)
}

func (qw *QueueWorker) processQueue() {
	// Find pending queue items that are ready to be processed
	query := `
		SELECT id, payload, attempts
		FROM task_queue 
		WHERE status = 'pending' 
		AND scheduled_for <= datetime('now')
		AND queue_type = 'generate_scheduled_tasks'
		ORDER BY scheduled_for ASC
		LIMIT 10
	`

	rows, err := qw.db.Query(query)
	if err != nil {
		log.Printf("Failed to query queue items: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var queueID, payloadJSON string
		var attempts int

		err := rows.Scan(&queueID, &payloadJSON, &attempts)
		if err != nil {
			log.Printf("Failed to scan queue item: %v", err)
			continue
		}

		qw.processQueueItem(queueID, payloadJSON, attempts)
	}
}

func (qw *QueueWorker) processQueueItem(queueID, payloadJSON string, attempts int) {
	log.Printf("Processing queue item %s (attempt %d)", queueID, attempts+1)

	// Mark as processing
	_, err := qw.db.Exec(
		"UPDATE task_queue SET status = 'processing', attempts = attempts + 1 WHERE id = ?",
		queueID,
	)
	if err != nil {
		log.Printf("Failed to mark queue item as processing: %v", err)
		return
	}

	// Parse payload
	var payload QueuePayload
	err = json.Unmarshal([]byte(payloadJSON), &payload)
	if err != nil {
		qw.markQueueItemFailed(queueID, fmt.Sprintf("Invalid payload JSON: %v", err))
		return
	}

	// Generate the task
	err = qw.generateScheduledTask(payload.ScheduleID, payload.TargetDate)
	if err != nil {
		// Check if we should retry
		maxAttempts := 3
		if attempts+1 >= maxAttempts {
			qw.markQueueItemFailed(queueID, err.Error())
		} else {
			// Reset to pending for retry
			qw.markQueueItemPending(queueID, err.Error())
		}
		return
	}

	// Mark as completed
	_, err = qw.db.Exec(
		"UPDATE task_queue SET status = 'completed', processed_at = datetime('now') WHERE id = ?",
		queueID,
	)
	if err != nil {
		log.Printf("Failed to mark queue item as completed: %v", err)
	}

	log.Printf("Successfully processed queue item %s", queueID)
}

func (qw *QueueWorker) markQueueItemFailed(queueID, errorMsg string) {
	_, err := qw.db.Exec(
		"UPDATE task_queue SET status = 'failed', error_message = ?, processed_at = datetime('now') WHERE id = ?",
		errorMsg, queueID,
	)
	if err != nil {
		log.Printf("Failed to mark queue item as failed: %v", err)
	}
	log.Printf("Queue item %s marked as failed: %s", queueID, errorMsg)
}

func (qw *QueueWorker) markQueueItemPending(queueID, errorMsg string) {
	_, err := qw.db.Exec(
		"UPDATE task_queue SET status = 'pending', error_message = ? WHERE id = ?",
		errorMsg, queueID,
	)
	if err != nil {
		log.Printf("Failed to mark queue item as pending for retry: %v", err)
	}
	log.Printf("Queue item %s marked as pending for retry: %s", queueID, errorMsg)
}

func (qw *QueueWorker) generateScheduledTask(scheduleID, targetDateStr string) error {
	// Parse target date
	targetDate, err := time.Parse("2006-01-02", targetDateStr)
	if err != nil {
		return fmt.Errorf("invalid target date format: %v", err)
	}

	// Get the schedule
	schedule, err := qw.getTaskSchedule(scheduleID)
	if err != nil {
		return fmt.Errorf("failed to get schedule: %v", err)
	}

	// Check if this day of week matches the schedule
	weekday := targetDate.Weekday().String()
	dayMatches := false
	for _, day := range schedule.DaysOfWeek {
		if weekday == day {
			dayMatches = true
			break
		}
	}

	if !dayMatches {
		log.Printf("Target date %s (%s) doesn't match schedule days: %v", targetDateStr, weekday, schedule.DaysOfWeek)
		return nil // Not an error, just skip this day
	}

	// Check if task already exists for this schedule and date
	existingTask, err := qw.checkExistingTask(scheduleID, targetDateStr)
	if err != nil {
		return fmt.Errorf("failed to check existing task: %v", err)
	}
	if existingTask {
		log.Printf("Task already exists for schedule %s on %s", scheduleID, targetDateStr)
		return nil // Task already exists, skip
	}

	// Create the task
	err = qw.createTaskFromSchedule(schedule, targetDate)
	if err != nil {
		return fmt.Errorf("failed to create task: %v", err)
	}

	log.Printf("Created task for schedule %s on %s", scheduleID, targetDateStr)
	return nil
}

func (qw *QueueWorker) getTaskSchedule(scheduleID string) (*TaskSchedule, error) {
	query := `
		SELECT id, family_id, created_by, title, description, task_type,
			   assigned_to, days_of_week, time_of_day, priority, points
		FROM task_schedules 
		WHERE id = ? AND active = true
	`

	var schedule TaskSchedule
	var assignedTo sql.NullString
	var timeOfDay sql.NullString
	var daysOfWeekJSON string

	err := qw.db.QueryRow(query, scheduleID).Scan(
		&schedule.ID,
		&schedule.FamilyID,
		&schedule.CreatedBy,
		&schedule.Title,
		&schedule.Description,
		&schedule.TaskType,
		&assignedTo,
		&daysOfWeekJSON,
		&timeOfDay,
		&schedule.Priority,
		&schedule.Points,
	)

	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if assignedTo.Valid {
		schedule.AssignedTo = &assignedTo.String
	}
	if timeOfDay.Valid {
		schedule.TimeOfDay = &timeOfDay.String
	}

	// Parse days of week JSON
	err = json.Unmarshal([]byte(daysOfWeekJSON), &schedule.DaysOfWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to parse days_of_week: %v", err)
	}

	return &schedule, nil
}

func (qw *QueueWorker) checkExistingTask(scheduleID, targetDate string) (bool, error) {
	query := `
		SELECT COUNT(*) FROM tasks 
		WHERE schedule_id = ? 
		AND DATE(created_at) = ?
	`

	var count int
	err := qw.db.QueryRow(query, scheduleID, targetDate).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (qw *QueueWorker) createTaskFromSchedule(schedule *TaskSchedule, targetDate time.Time) error {
	// Handle assigned_to value
	var assignedToValue interface{}
	if schedule.AssignedTo != nil {
		assignedToValue = *schedule.AssignedTo
	} else {
		assignedToValue = nil
	}

	// Create due date with optional time
	var dueDate *time.Time
	if schedule.TimeOfDay != nil {
		// Parse time and set it on the target date
		timeStr := *schedule.TimeOfDay
		if timePart, err := time.Parse("15:04", timeStr); err == nil {
			dueDateWithTime := time.Date(
				targetDate.Year(), targetDate.Month(), targetDate.Day(),
				timePart.Hour(), timePart.Minute(), 0, 0, targetDate.Location(),
			)
			dueDate = &dueDateWithTime
		}
	}

	var dueDateValue interface{}
	if dueDate != nil {
		dueDateValue = dueDate.Format("2006-01-02 15:04:05")
	} else {
		dueDateValue = nil
	}

	query := `
		INSERT INTO tasks (family_id, assigned_to, title, description, task_type,
						  status, priority, points, due_date, created_by, schedule_id)
		VALUES (?, ?, ?, ?, ?, 'pending', ?, ?, ?, ?, ?)
		RETURNING id
	`

	var newID string
	err := qw.db.QueryRow(query,
		schedule.FamilyID,
		assignedToValue,
		schedule.Title,
		schedule.Description,
		schedule.TaskType,
		schedule.Priority,
		schedule.Points,
		dueDateValue,
		schedule.CreatedBy,
		schedule.ID,
	).Scan(&newID)

	return err
}
