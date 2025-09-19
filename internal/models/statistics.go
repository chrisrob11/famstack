package models

// FamilyStatistics represents statistics for a family
type FamilyStatistics struct {
	FamilyID       string  `json:"family_id"`
	MemberCount    int     `json:"member_count"`
	TotalTasks     int     `json:"total_tasks"`
	CompletedTasks int     `json:"completed_tasks"`
	PendingTasks   int     `json:"pending_tasks"`
	CompletionRate float64 `json:"completion_rate"` // Percentage (0-100)
}
