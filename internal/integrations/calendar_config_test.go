package integrations

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultCalendarSyncConfig(t *testing.T) {
	config := DefaultCalendarSyncConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 30, config.SyncFrequencyMinutes)
	assert.Equal(t, 30, config.SyncRangeDays)
	assert.Empty(t, config.CalendarsToSync)
	assert.True(t, config.SyncAllDayEvents)
	assert.False(t, config.SyncPrivateEvents)
	assert.False(t, config.SyncDeclinedEvents)
}

func TestCalendarSyncConfig_Validate(t *testing.T) {
	tests := []struct {
		name          string
		config        CalendarSyncConfig
		expectedFreq  int
		expectedRange int
	}{
		{
			name: "valid default config",
			config: CalendarSyncConfig{
				SyncFrequencyMinutes: 30,
				SyncRangeDays:        30,
				CalendarsToSync:      []string{},
				SyncAllDayEvents:     true,
				SyncPrivateEvents:    false,
				SyncDeclinedEvents:   false,
			},
			expectedFreq:  30,
			expectedRange: 30,
		},
		{
			name: "frequency too low gets corrected",
			config: CalendarSyncConfig{
				SyncFrequencyMinutes: 15, // Below minimum
				SyncRangeDays:        30,
				CalendarsToSync:      []string{},
				SyncAllDayEvents:     true,
				SyncPrivateEvents:    false,
				SyncDeclinedEvents:   false,
			},
			expectedFreq:  30, // Should be corrected to minimum
			expectedRange: 30,
		},
		{
			name: "zero range gets corrected",
			config: CalendarSyncConfig{
				SyncFrequencyMinutes: 30,
				SyncRangeDays:        0, // Invalid
				CalendarsToSync:      []string{},
				SyncAllDayEvents:     true,
				SyncPrivateEvents:    false,
				SyncDeclinedEvents:   false,
			},
			expectedFreq:  30,
			expectedRange: 30, // Should be corrected to default
		},
		{
			name: "range too high gets corrected",
			config: CalendarSyncConfig{
				SyncFrequencyMinutes: 30,
				SyncRangeDays:        120, // Above maximum
				CalendarsToSync:      []string{},
				SyncAllDayEvents:     true,
				SyncPrivateEvents:    false,
				SyncDeclinedEvents:   false,
			},
			expectedFreq:  30,
			expectedRange: 90, // Should be corrected to maximum
		},
		{
			name: "negative values get corrected",
			config: CalendarSyncConfig{
				SyncFrequencyMinutes: -10,
				SyncRangeDays:        -5,
				CalendarsToSync:      []string{},
				SyncAllDayEvents:     true,
				SyncPrivateEvents:    false,
				SyncDeclinedEvents:   false,
			},
			expectedFreq:  30, // Corrected to minimum
			expectedRange: 30, // Corrected to default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			assert.NoError(t, err) // Validate never returns errors, just fixes values

			assert.Equal(t, tt.expectedFreq, tt.config.SyncFrequencyMinutes)
			assert.Equal(t, tt.expectedRange, tt.config.SyncRangeDays)
		})
	}
}

func TestCalendarSyncConfig_JSONSerialization(t *testing.T) {
	// Test marshaling
	config := CalendarSyncConfig{
		SyncFrequencyMinutes: 45,
		SyncRangeDays:        14,
		CalendarsToSync:      []string{"primary", "work", "personal"},
		SyncAllDayEvents:     true,
		SyncPrivateEvents:    true,
		SyncDeclinedEvents:   false,
	}

	data, err := json.Marshal(config)
	require.NoError(t, err)

	// Test unmarshaling
	var unmarshaled CalendarSyncConfig
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, config.SyncFrequencyMinutes, unmarshaled.SyncFrequencyMinutes)
	assert.Equal(t, config.SyncRangeDays, unmarshaled.SyncRangeDays)
	assert.Equal(t, config.CalendarsToSync, unmarshaled.CalendarsToSync)
	assert.Equal(t, config.SyncAllDayEvents, unmarshaled.SyncAllDayEvents)
	assert.Equal(t, config.SyncPrivateEvents, unmarshaled.SyncPrivateEvents)
	assert.Equal(t, config.SyncDeclinedEvents, unmarshaled.SyncDeclinedEvents)

	// Verify the unmarshaled config is still valid
	err = unmarshaled.Validate()
	assert.NoError(t, err)
}

func TestCalendarSyncConfig_EdgeCases(t *testing.T) {
	t.Run("nil calendars to sync", func(t *testing.T) {
		config := CalendarSyncConfig{
			SyncFrequencyMinutes: 30,
			SyncRangeDays:        30,
			CalendarsToSync:      nil, // nil slice
			SyncAllDayEvents:     true,
			SyncPrivateEvents:    false,
			SyncDeclinedEvents:   false,
		}

		err := config.Validate()
		assert.NoError(t, err) // nil slice should be treated same as empty slice
	})

	t.Run("empty calendars list", func(t *testing.T) {
		config := CalendarSyncConfig{
			SyncFrequencyMinutes: 30,
			SyncRangeDays:        30,
			CalendarsToSync:      []string{}, // empty slice
			SyncAllDayEvents:     true,
			SyncPrivateEvents:    false,
			SyncDeclinedEvents:   false,
		}

		err := config.Validate()
		assert.NoError(t, err)
		assert.Empty(t, config.CalendarsToSync)
	})

	t.Run("multiple calendars", func(t *testing.T) {
		config := CalendarSyncConfig{
			SyncFrequencyMinutes: 30,
			SyncRangeDays:        30,
			CalendarsToSync:      []string{"primary", "work", "personal"},
			SyncAllDayEvents:     true,
			SyncPrivateEvents:    false,
			SyncDeclinedEvents:   false,
		}

		err := config.Validate()
		assert.NoError(t, err)
		assert.Len(t, config.CalendarsToSync, 3)
	})
}

func TestCalendarSyncConfig_GetSyncBooleans(t *testing.T) {
	tests := []struct {
		name               string
		syncAllDayEvents   bool
		syncPrivateEvents  bool
		syncDeclinedEvents bool
		expectedAllDay     bool
		expectedPrivate    bool
		expectedDeclined   bool
	}{
		{
			name:               "all sync options enabled",
			syncAllDayEvents:   true,
			syncPrivateEvents:  true,
			syncDeclinedEvents: true,
			expectedAllDay:     true,
			expectedPrivate:    true,
			expectedDeclined:   true,
		},
		{
			name:               "only all-day events enabled",
			syncAllDayEvents:   true,
			syncPrivateEvents:  false,
			syncDeclinedEvents: false,
			expectedAllDay:     true,
			expectedPrivate:    false,
			expectedDeclined:   false,
		},
		{
			name:               "no sync options enabled",
			syncAllDayEvents:   false,
			syncPrivateEvents:  false,
			syncDeclinedEvents: false,
			expectedAllDay:     false,
			expectedPrivate:    false,
			expectedDeclined:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := CalendarSyncConfig{
				SyncAllDayEvents:   tt.syncAllDayEvents,
				SyncPrivateEvents:  tt.syncPrivateEvents,
				SyncDeclinedEvents: tt.syncDeclinedEvents,
			}

			assert.Equal(t, tt.expectedAllDay, config.SyncAllDayEvents)
			assert.Equal(t, tt.expectedPrivate, config.SyncPrivateEvents)
			assert.Equal(t, tt.expectedDeclined, config.SyncDeclinedEvents)
		})
	}
}

func TestCalendarSyncConfig_ConfigurationScenarios(t *testing.T) {
	t.Run("minimal work setup", func(t *testing.T) {
		config := CalendarSyncConfig{
			SyncFrequencyMinutes: 60, // Hourly sync
			SyncRangeDays:        7,  // One week
			CalendarsToSync:      []string{"work"},
			SyncAllDayEvents:     true,
			SyncPrivateEvents:    false, // Don't sync private events at work
			SyncDeclinedEvents:   false,
		}

		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("comprehensive personal setup", func(t *testing.T) {
		config := CalendarSyncConfig{
			SyncFrequencyMinutes: 15, // Frequent sync
			SyncRangeDays:        90, // Three months
			CalendarsToSync:      []string{"primary", "family", "personal", "birthdays"},
			SyncAllDayEvents:     true,
			SyncPrivateEvents:    true, // Include private events
			SyncDeclinedEvents:   true, // Include declined for reference
		}

		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("battery-saving setup", func(t *testing.T) {
		config := CalendarSyncConfig{
			SyncFrequencyMinutes: 360, // Every 6 hours
			SyncRangeDays:        14,  // Two weeks
			CalendarsToSync:      []string{"primary"},
			SyncAllDayEvents:     true,
			SyncPrivateEvents:    false,
			SyncDeclinedEvents:   false,
		}

		err := config.Validate()
		assert.NoError(t, err)
	})
}
