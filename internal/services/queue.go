package services

import (
	"encoding/json"
	"fmt"
	"time"

	"famstack/internal/database"
)

type QueueService struct {
	db *database.DB
}

func NewQueueService(db *database.DB) *QueueService {
	return &QueueService{
		db: db,
	}
}

// QueueTaskGeneration queues task generation for a schedule over the next 14 days
func (qs *QueueService) QueueTaskGeneration(scheduleID string) error {
	now := time.Now()

	// Generate queue items for the next 14 days
	for i := range 14 {
		targetDate := now.AddDate(0, 0, i)
		payload := map[string]any{
			"schedule_id": scheduleID,
			"target_date": targetDate.Format("2006-01-02"),
		}

		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %v", err)
		}

		// Schedule processing for midnight of the target date minus 1 day
		scheduledFor := targetDate.AddDate(0, 0, -1).Truncate(24 * time.Hour)
		if scheduledFor.Before(now) {
			scheduledFor = now // Process immediately if it's in the past
		}

		query := `
			INSERT OR IGNORE INTO task_queue (queue_type, payload, scheduled_for)
			VALUES ('generate_scheduled_tasks', ?, ?)
		`

		_, err = qs.db.Exec(query, string(payloadJSON), scheduledFor.Format("2006-01-02 15:04:05"))
		if err != nil {
			return fmt.Errorf("failed to insert queue item: %v", err)
		}
	}

	return nil
}

// QueueTaskGenerationUpdate removes old queue items and queues new ones for a schedule
func (qs *QueueService) QueueTaskGenerationUpdate(scheduleID string) error {
	// Remove existing unprocessed queue items for this schedule
	deleteQuery := `
		DELETE FROM task_queue 
		WHERE queue_type = 'generate_scheduled_tasks' 
		AND status = 'pending'
		AND json_extract(payload, '$.schedule_id') = ?
	`

	_, err := qs.db.Exec(deleteQuery, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to delete existing queue items: %v", err)
	}

	// Queue new items
	return qs.QueueTaskGeneration(scheduleID)
}

// RemoveScheduleFromQueue removes all queue items for a deleted schedule
func (qs *QueueService) RemoveScheduleFromQueue(scheduleID string) error {
	deleteQuery := `
		DELETE FROM task_queue 
		WHERE queue_type = 'generate_scheduled_tasks' 
		AND status = 'pending'
		AND json_extract(payload, '$.schedule_id') = ?
	`

	_, err := qs.db.Exec(deleteQuery, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to remove schedule from queue: %v", err)
	}

	return nil
}
