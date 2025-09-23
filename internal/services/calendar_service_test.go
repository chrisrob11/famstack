package services

import (
	"fmt"
	"os"
	"testing"
	"time"

	"famstack/internal/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *database.Fascade {
	dbFile := fmt.Sprintf("test_db_%d.db", time.Now().UnixNano())
	db, err := database.New(dbFile)
	require.NoError(t, err)

	err = db.MigrateUp()
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
		os.Remove(dbFile)
	})

	return db
}

func TestGetUnifiedCalendarEvents_TimezoneConversion(t *testing.T) {
	db := setupTestDB(t)
	service := NewCalendarService(db)

	// 1. Seed Data
	familyID := "fam_tz_test"
	timezone := "America/New_York"
	_, err := db.Exec(`INSERT INTO families (id, name, timezone) VALUES (?, ?, ?)`, familyID, "Timezone Test Family", timezone)
	require.NoError(t, err)

	eventID := "event_tz_test"
	// This time is 1:00 PM UTC on the test date.
	// In America/New_York (UTC-4), this should be 9:00 AM.
	utcStartTime := time.Date(2025, 9, 23, 13, 0, 0, 0, time.UTC)
	utcEndTime := utcStartTime.Add(1 * time.Hour)

	_, err = db.Exec(`
		INSERT INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		eventID, familyID, "Timezone Conversion Test Event", utcStartTime, utcEndTime, false, "event", "user_test", time.Now(), time.Now(),
	)
	require.NoError(t, err)

	// 2. Call the service method
	// The date range needs to be converted to UTC for the query
	loc, err := time.LoadLocation(timezone)
	require.NoError(t, err)
	rangeStart := time.Date(2025, 9, 23, 0, 0, 0, 0, loc)
	rangeEnd := rangeStart.Add(24 * time.Hour)

	events, err := service.GetUnifiedCalendarEvents(familyID, rangeStart, rangeEnd)
	require.NoError(t, err)
	require.Len(t, events, 1, "Expected to find one event")

	// 3. Assert the results
	event := events[0]
	assert.Equal(t, eventID, event.ID)

	// Check the location of the time object
	expectedLocation, err := time.LoadLocation(timezone)
	require.NoError(t, err)
	assert.Equal(t, expectedLocation, event.StartTime.Location(), "StartTime should have the family's timezone location")
	assert.Equal(t, expectedLocation, event.EndTime.Location(), "EndTime should have the family's timezone location")

	// Check the actual time
	// 13:00 UTC should be 9:00 AM in New York (EDT is UTC-4)
	assert.Equal(t, 9, event.StartTime.Hour(), "Hour should be 9 AM in New York time")
	assert.Equal(t, 0, event.StartTime.Minute(), "Minute should be 00")

	// Check the formatted string for the offset
	// The offset for America/New_York in September is -04:00
	assert.Contains(t, event.StartTime.Format(time.RFC3339), "-04:00", "Formatted time string should include the -04:00 offset")
}

func TestTimezoneConversions(t *testing.T) {
	// Create a mock calendar service for testing
	service := &CalendarService{}

	tests := []struct {
		name           string
		inputTime      time.Time
		familyTimezone string
		expectError    bool
	}{
		{
			name:           "UTC to UTC should remain unchanged",
			inputTime:      time.Date(2023, 9, 22, 15, 30, 0, 0, time.UTC),
			familyTimezone: "UTC",
			expectError:    false,
		},
		{
			name:           "Naive time in EST timezone",
			inputTime:      time.Date(2023, 9, 22, 15, 30, 0, 0, time.UTC), // Treating as naive
			familyTimezone: "America/New_York",
			expectError:    false,
		},
		{
			name:           "Naive time in PST timezone",
			inputTime:      time.Date(2023, 9, 22, 15, 30, 0, 0, time.UTC), // Treating as naive
			familyTimezone: "America/Los_Angeles",
			expectError:    false,
		},
		{
			name:           "Invalid timezone should error",
			inputTime:      time.Date(2023, 9, 22, 15, 30, 0, 0, time.UTC),
			familyTimezone: "Invalid/Timezone",
			expectError:    true,
		},
		{
			name:           "Empty timezone defaults to UTC",
			inputTime:      time.Date(2023, 9, 22, 15, 30, 0, 0, time.UTC),
			familyTimezone: "",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test conversion to UTC
			utcTime, err := service.convertToUTC(tt.inputTime, tt.familyTimezone)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// UTC time should always be in UTC location
			if utcTime.Location() != time.UTC {
				t.Errorf("Expected UTC location, got %v", utcTime.Location())
			}

			// Test round trip: convert back from UTC
			localTime, err := service.convertFromUTC(utcTime, tt.familyTimezone)
			if err != nil {
				t.Errorf("Unexpected error converting from UTC: %v", err)
				return
			}

			// For UTC timezone, times should be identical
			if tt.familyTimezone == "UTC" || tt.familyTimezone == "" {
				if !utcTime.Equal(localTime) {
					t.Errorf("UTC round trip failed: %v != %v", utcTime, localTime)
				}
				return
			}

			// For other timezones, verify the time components are preserved
			expectedLoc, _ := time.LoadLocation(tt.familyTimezone)
			if localTime.Location().String() != expectedLoc.String() {
				t.Errorf("Expected location %v, got %v", expectedLoc, localTime.Location())
			}
		})
	}
}

func TestTimezoneRoundTripWithSpecificTimes(t *testing.T) {
	service := &CalendarService{}

	// Test specific scenarios that are important for calendar applications
	testCases := []struct {
		name           string
		localHour      int
		localMinute    int
		familyTimezone string
		date           time.Time
	}{
		{
			name:           "9 AM Eastern Standard Time",
			localHour:      9,
			localMinute:    0,
			familyTimezone: "America/New_York",
			date:           time.Date(2023, 12, 15, 0, 0, 0, 0, time.UTC), // Winter (EST)
		},
		{
			name:           "9 AM Eastern Daylight Time",
			localHour:      9,
			localMinute:    0,
			familyTimezone: "America/New_York",
			date:           time.Date(2023, 7, 15, 0, 0, 0, 0, time.UTC), // Summer (EDT)
		},
		{
			name:           "3 PM Pacific Standard Time",
			localHour:      15,
			localMinute:    30,
			familyTimezone: "America/Los_Angeles",
			date:           time.Date(2023, 12, 15, 0, 0, 0, 0, time.UTC), // Winter (PST)
		},
		{
			name:           "3 PM Pacific Daylight Time",
			localHour:      15,
			localMinute:    30,
			familyTimezone: "America/Los_Angeles",
			date:           time.Date(2023, 7, 15, 0, 0, 0, 0, time.UTC), // Summer (PDT)
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a naive time (as if entered by user in their local timezone)
			naiveTime := time.Date(tc.date.Year(), tc.date.Month(), tc.date.Day(),
				tc.localHour, tc.localMinute, 0, 0, time.UTC)

			// Convert to UTC for storage
			utcTime, err := service.convertToUTC(naiveTime, tc.familyTimezone)
			if err != nil {
				t.Errorf("Error converting to UTC: %v", err)
				return
			}

			// Convert back for display
			displayTime, err := service.convertFromUTC(utcTime, tc.familyTimezone)
			if err != nil {
				t.Errorf("Error converting from UTC: %v", err)
				return
			}

			// The display time should have the same hour and minute as originally entered
			if displayTime.Hour() != tc.localHour || displayTime.Minute() != tc.localMinute {
				t.Errorf("Round trip failed: expected %02d:%02d, got %02d:%02d",
					tc.localHour, tc.localMinute, displayTime.Hour(), displayTime.Minute())
			}

			t.Logf("Original: %02d:%02d in %s", tc.localHour, tc.localMinute, tc.familyTimezone)
			t.Logf("UTC storage: %s", utcTime.Format("2006-01-02 15:04:05 MST"))
			t.Logf("Display: %s", displayTime.Format("2006-01-02 15:04:05 MST"))
		})
	}
}

func TestTimezoneEdgeCases(t *testing.T) {
	service := &CalendarService{}

	// Test daylight saving time transitions
	t.Run("DST Spring Forward", func(t *testing.T) {
		// 2023 DST begins March 12, 2:00 AM -> 3:00 AM in US Eastern
		// Test a time that would be in the "gap"
		gapTime := time.Date(2023, 3, 12, 2, 30, 0, 0, time.UTC) // 2:30 AM (doesn't exist)

		utcTime, err := service.convertToUTC(gapTime, "America/New_York")
		if err != nil {
			t.Errorf("Error handling DST gap: %v", err)
			return
		}

		displayTime, err := service.convertFromUTC(utcTime, "America/New_York")
		if err != nil {
			t.Errorf("Error converting back from DST gap: %v", err)
		}

		t.Logf("DST gap handling: input 2:30 AM -> UTC %s -> display %s",
			utcTime.Format("15:04:05 MST"), displayTime.Format("15:04:05 MST"))
	})

	t.Run("DST Fall Back", func(t *testing.T) {
		// 2023 DST ends November 5, 2:00 AM -> 1:00 AM in US Eastern
		// Test a time that would be ambiguous
		ambiguousTime := time.Date(2023, 11, 5, 1, 30, 0, 0, time.UTC) // 1:30 AM (exists twice)

		utcTime, err := service.convertToUTC(ambiguousTime, "America/New_York")
		if err != nil {
			t.Errorf("Error handling DST ambiguity: %v", err)
			return
		}

		displayTime, err := service.convertFromUTC(utcTime, "America/New_York")
		if err != nil {
			t.Errorf("Error converting back from DST ambiguity: %v", err)
		}

		t.Logf("DST ambiguity handling: input 1:30 AM -> UTC %s -> display %s",
			utcTime.Format("15:04:05 MST"), displayTime.Format("15:04:05 MST"))
	})
}
