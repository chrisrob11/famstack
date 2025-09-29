package integrations

// CalendarSyncConfig represents configuration for calendar sync integrations
// Stored in integrations.settings as JSON, identified by settings_type="CalendarSyncConfig"
type CalendarSyncConfig struct {
	// Sync behavior
	SyncFrequencyMinutes int      `json:"sync_frequency_minutes"` // Minimum 30 minutes
	SyncRangeDays        int      `json:"sync_range_days"`        // Fixed at 30 days for now
	CalendarsToSync      []string `json:"calendars_to_sync"`      // Calendar IDs to sync, empty = all

	// Sync preferences
	SyncAllDayEvents   bool `json:"sync_all_day_events"`  // Include all-day events
	SyncPrivateEvents  bool `json:"sync_private_events"`  // Include private events
	SyncDeclinedEvents bool `json:"sync_declined_events"` // Include events user declined
}

// DefaultCalendarSyncConfig returns sensible defaults for new calendar integrations
func DefaultCalendarSyncConfig() *CalendarSyncConfig {
	return &CalendarSyncConfig{
		SyncFrequencyMinutes: 30,
		SyncRangeDays:        30,
		CalendarsToSync:      []string{}, // Empty = sync all calendars
		SyncAllDayEvents:     true,
		SyncPrivateEvents:    false, // Default to not syncing private events
		SyncDeclinedEvents:   false, // Default to not syncing declined events
	}
}

// Validate checks if the configuration is valid and applies constraints
func (c *CalendarSyncConfig) Validate() error {
	if c.SyncFrequencyMinutes < 30 {
		c.SyncFrequencyMinutes = 30 // Enforce minimum 30 minutes
	}

	if c.SyncRangeDays <= 0 {
		c.SyncRangeDays = 30 // Default to 30 days
	}

	if c.SyncRangeDays > 90 {
		c.SyncRangeDays = 90 // Max 90 days for now
	}

	return nil
}

// GetConfigType returns the settings_type value for this config
func (c *CalendarSyncConfig) GetConfigType() string {
	return "CalendarSyncConfig"
}
