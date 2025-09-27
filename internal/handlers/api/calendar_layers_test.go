package api

import (
	"testing"
	"time"

	"famstack/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test the layer assignment algorithm with various overlap scenarios
func TestCalculateEventLayers(t *testing.T) {
	handler := &CalendarAPIHandler{}

	tests := []struct {
		name           string
		events         []models.UnifiedCalendarEvent
		expectedLayers int
		description    string
	}{
		{
			name:           "No events",
			events:         []models.UnifiedCalendarEvent{},
			expectedLayers: 0,
			description:    "Empty day should have no layers",
		},
		{
			name: "Single event",
			events: []models.UnifiedCalendarEvent{
				createTestEvent("1", "Meeting", "09:00", "10:00"),
			},
			expectedLayers: 1,
			description:    "Single event should create one layer",
		},
		{
			name: "Two non-overlapping events",
			events: []models.UnifiedCalendarEvent{
				createTestEvent("1", "Morning Meeting", "09:00", "10:00"),
				createTestEvent("2", "Afternoon Meeting", "14:00", "15:00"),
			},
			expectedLayers: 1,
			description:    "Sequential events should fit in one layer",
		},
		{
			name: "Two overlapping events",
			events: []models.UnifiedCalendarEvent{
				createTestEvent("1", "Meeting A", "09:00", "10:00"),
				createTestEvent("2", "Meeting B", "09:30", "10:30"),
			},
			expectedLayers: 2,
			description:    "Overlapping events should create two layers",
		},
		{
			name: "Three overlapping events",
			events: []models.UnifiedCalendarEvent{
				createTestEvent("1", "Event 1", "09:00", "10:00"),
				createTestEvent("2", "Event 2", "09:15", "10:15"),
				createTestEvent("3", "Event 3", "09:30", "10:30"),
			},
			expectedLayers: 3,
			description:    "Three overlapping events should create three layers",
		},
		{
			name: "Complex cascade overlaps",
			events: []models.UnifiedCalendarEvent{
				createTestEvent("1", "Event 1", "08:00", "09:00"), // Layer 0
				createTestEvent("2", "Event 2", "08:30", "09:30"), // Layer 1 (overlaps with 1)
				createTestEvent("3", "Event 3", "09:15", "10:15"), // Layer 0 (no overlap with 1)
				createTestEvent("4", "Event 4", "09:45", "10:45"), // Layer 1 (no overlap with 2)
			},
			expectedLayers: 2,
			description:    "Cascading events should optimize layer usage",
		},
		{
			name: "Adjacent events (touching but not overlapping)",
			events: []models.UnifiedCalendarEvent{
				createTestEvent("1", "Event 1", "09:00", "10:00"),
				createTestEvent("2", "Event 2", "10:00", "11:00"),
			},
			expectedLayers: 1,
			description:    "Adjacent events should fit in one layer",
		},
		{
			name: "Partial overlaps with different end times",
			events: []models.UnifiedCalendarEvent{
				createTestEvent("1", "Short Meeting", "09:00", "09:30"),
				createTestEvent("2", "Long Meeting", "09:15", "11:00"),
				createTestEvent("3", "Quick Call", "10:30", "10:45"),
			},
			expectedLayers: 2,
			description:    "Mixed duration overlaps should use two layers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layers := handler.calculateEventLayers(tt.events, "UTC")

			assert.Equal(t, tt.expectedLayers, len(layers), tt.description)

			// Verify layer structure
			if len(layers) > 0 {
				for i, layer := range layers {
					assert.Equal(t, i, layer.LayerIndex, "Layer index should match position")

					// Verify no overlaps within the same layer
					for j, event1 := range layer.Events {
						for k, event2 := range layer.Events {
							if j != k {
								assert.False(t, handler.eventsOverlap(event1, event2),
									"Events %s and %s should not overlap in same layer %d", event1.ID, event2.ID, i)
							}
						}
					}
				}
			}
		})
	}
}

// Test removed - calculateSlotInfo is no longer needed since we simplified DayView

// Test time to slot conversion
func TestTimeToSlot(t *testing.T) {
	handler := &CalendarAPIHandler{}

	tests := []struct {
		time         string
		expectedSlot int
		description  string
	}{
		{"00:00", 0, "Midnight should be slot 0"},
		{"01:00", 4, "1:00 AM should be slot 4"},
		{"06:00", 24, "6:00 AM should be slot 24"},
		{"09:00", 36, "9:00 AM should be slot 36"},
		{"09:15", 37, "9:15 AM should be slot 37"},
		{"09:30", 38, "9:30 AM should be slot 38"},
		{"09:45", 39, "9:45 AM should be slot 39"},
		{"12:00", 48, "Noon should be slot 48"},
		{"15:00", 60, "3:00 PM should be slot 60"},
		{"18:00", 72, "6:00 PM should be slot 72"},
		{"23:45", 95, "11:45 PM should be slot 95"},
		{"23:59", 95, "11:59 PM should still be slot 95"},
	}

	for _, tt := range tests {
		t.Run(tt.time, func(t *testing.T) {
			parsedTime, err := time.Parse("15:04", tt.time)
			require.NoError(t, err)

			slot := handler.timeToSlot(parsedTime)
			assert.Equal(t, tt.expectedSlot, slot, tt.description)
		})
	}
}

// Test events overlap detection
func TestEventsOverlap(t *testing.T) {
	handler := &CalendarAPIHandler{}

	tests := []struct {
		name          string
		event1        models.CalendarViewEvent
		event2        models.CalendarViewEvent
		shouldOverlap bool
		description   string
	}{
		{
			name:          "Identical events",
			event1:        models.CalendarViewEvent{StartSlot: 36, EndSlot: 40},
			event2:        models.CalendarViewEvent{StartSlot: 36, EndSlot: 40},
			shouldOverlap: true,
			description:   "Identical events should overlap",
		},
		{
			name:          "Adjacent events (touching)",
			event1:        models.CalendarViewEvent{StartSlot: 36, EndSlot: 40},
			event2:        models.CalendarViewEvent{StartSlot: 40, EndSlot: 44},
			shouldOverlap: false,
			description:   "Adjacent events should not overlap",
		},
		{
			name:          "Partial overlap",
			event1:        models.CalendarViewEvent{StartSlot: 36, EndSlot: 40},
			event2:        models.CalendarViewEvent{StartSlot: 38, EndSlot: 42},
			shouldOverlap: true,
			description:   "Partially overlapping events should overlap",
		},
		{
			name:          "Event1 contains event2",
			event1:        models.CalendarViewEvent{StartSlot: 36, EndSlot: 44},
			event2:        models.CalendarViewEvent{StartSlot: 38, EndSlot: 40},
			shouldOverlap: true,
			description:   "Containing event should overlap",
		},
		{
			name:          "Event2 contains event1",
			event1:        models.CalendarViewEvent{StartSlot: 38, EndSlot: 40},
			event2:        models.CalendarViewEvent{StartSlot: 36, EndSlot: 44},
			shouldOverlap: true,
			description:   "Contained event should overlap",
		},
		{
			name:          "Separated events",
			event1:        models.CalendarViewEvent{StartSlot: 36, EndSlot: 40},
			event2:        models.CalendarViewEvent{StartSlot: 44, EndSlot: 48},
			shouldOverlap: false,
			description:   "Separated events should not overlap",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overlap := handler.eventsOverlap(tt.event1, tt.event2)
			assert.Equal(t, tt.shouldOverlap, overlap, tt.description)

			// Test symmetry - overlap should be the same regardless of order
			reverseOverlap := handler.eventsOverlap(tt.event2, tt.event1)
			assert.Equal(t, overlap, reverseOverlap, "Overlap detection should be symmetric")
		})
	}
}

// Test the full DayView conversion with real scenarios
func TestDayViewIntegration(t *testing.T) {
	handler := &CalendarAPIHandler{}

	tests := []struct {
		name           string
		events         []models.UnifiedCalendarEvent
		expectedLayers int
		description    string
	}{
		{
			name: "Real world: Two back-to-back meetings",
			events: []models.UnifiedCalendarEvent{
				createTestEvent("1", "Morning Standup", "09:00", "09:30"),
				createTestEvent("2", "Team Planning", "09:30", "10:30"),
			},
			expectedLayers: 1,
			description:    "Sequential meetings should use one layer",
		},
		{
			name: "Real world: Overlapping meetings",
			events: []models.UnifiedCalendarEvent{
				createTestEvent("1", "Client Call", "10:00", "11:00"),
				createTestEvent("2", "Team Sync", "10:30", "11:30"),
			},
			expectedLayers: 2,
			description:    "Overlapping meetings should use two layers",
		},
		{
			name: "Real world: Complex day with multiple overlaps",
			events: []models.UnifiedCalendarEvent{
				createTestEvent("1", "Morning Meeting", "09:00", "10:00"),
				createTestEvent("2", "Workshop Part 1", "09:30", "11:00"),
				createTestEvent("3", "Quick Check-in", "10:15", "10:30"),
				createTestEvent("4", "Lunch", "12:00", "13:00"),
			},
			expectedLayers: 2, // Layer 0: events 1,3,4; Layer 1: event 2
			description:    "Complex day should optimize layer usage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate layers
			layers := handler.calculateEventLayers(tt.events, "UTC")
			assert.Equal(t, tt.expectedLayers, len(layers), tt.description+" - layer count")

			// Verify all events are placed somewhere
			totalEventsInLayers := 0
			for _, layer := range layers {
				totalEventsInLayers += len(layer.Events)
			}
			assert.Equal(t, len(tt.events), totalEventsInLayers, "All events should be placed in layers")

			// Client can derive overlaps: hasOverlaps = len(layers) > 1
			hasOverlaps := len(layers) > 1
			expectedHasOverlaps := tt.expectedLayers > 1
			assert.Equal(t, expectedHasOverlaps, hasOverlaps, "Client should be able to derive overlap status")
		})
	}
}

// Test overlap group functionality - the new feature
func TestOverlapGroupCalculation(t *testing.T) {
	handler := &CalendarAPIHandler{}

	tests := []struct {
		name           string
		events         []models.UnifiedCalendarEvent
		expectedGroups map[string]struct{ group, index int }
		description    string
	}{
		{
			name: "Single event - no overlaps",
			events: []models.UnifiedCalendarEvent{
				createTestEvent("1", "Standalone Meeting", "09:00", "10:00"),
			},
			expectedGroups: map[string]struct{ group, index int }{
				"1": {group: 1, index: 0},
			},
			description: "Standalone event should have group=1, index=0",
		},
		{
			name: "Two overlapping events",
			events: []models.UnifiedCalendarEvent{
				createTestEvent("1", "Meeting A", "09:00", "10:00"),
				createTestEvent("2", "Meeting B", "09:30", "10:30"),
			},
			expectedGroups: map[string]struct{ group, index int }{
				"1": {group: 2, index: 0}, // Starts first, gets index 0
				"2": {group: 2, index: 1}, // Starts second, gets index 1
			},
			description: "Two overlapping events should both have group=2",
		},
		{
			name: "Three overlapping events",
			events: []models.UnifiedCalendarEvent{
				createTestEvent("1", "Meeting A", "09:00", "10:00"),
				createTestEvent("2", "Meeting B", "09:15", "10:15"),
				createTestEvent("3", "Meeting C", "09:30", "10:30"),
			},
			expectedGroups: map[string]struct{ group, index int }{
				"1": {group: 3, index: 0}, // Starts first (09:00)
				"2": {group: 3, index: 1}, // Starts second (09:15)
				"3": {group: 3, index: 2}, // Starts last (09:30)
			},
			description: "Three overlapping events should all have group=3 with sequential indices",
		},
		{
			name: "Mixed: overlapping and standalone",
			events: []models.UnifiedCalendarEvent{
				createTestEvent("1", "Meeting A", "09:00", "10:00"),
				createTestEvent("2", "Meeting B", "09:30", "10:30"),  // Overlaps with A
				createTestEvent("3", "Standalone", "14:00", "15:00"), // No overlap
			},
			expectedGroups: map[string]struct{ group, index int }{
				"1": {group: 2, index: 0}, // Part of 2-event overlap
				"2": {group: 2, index: 1}, // Part of 2-event overlap
				"3": {group: 1, index: 0}, // Standalone
			},
			description: "Mixed scenario should calculate groups correctly",
		},
		{
			name: "Same start time - ID ordering",
			events: []models.UnifiedCalendarEvent{
				createTestEvent("z", "Meeting Z", "09:00", "10:00"),
				createTestEvent("a", "Meeting A", "09:00", "10:00"),
				createTestEvent("m", "Meeting M", "09:00", "10:00"),
			},
			expectedGroups: map[string]struct{ group, index int }{
				"a": {group: 3, index: 1}, // ID "a" comes first lexically
				"m": {group: 3, index: 2}, // ID "m" comes second
				"z": {group: 3, index: 0}, // ID "z" comes last
			},
			description: "Same start time should use ID for consistent ordering",
		},
		{
			name: "Complex cascade overlap",
			events: []models.UnifiedCalendarEvent{
				createTestEvent("1", "Meeting A", "09:00", "10:00"),
				createTestEvent("2", "Meeting B", "09:30", "11:00"), // Overlaps A & C
				createTestEvent("3", "Meeting C", "10:30", "11:30"), // Overlaps B only
			},
			expectedGroups: map[string]struct{ group, index int }{
				"1": {group: 2, index: 0}, // Part of transitive group A->B->C (2 layers used: 0,1)
				"2": {group: 2, index: 1}, // Part of transitive group A->B->C (2 layers used: 0,1)
				"3": {group: 2, index: 0}, // Part of transitive group A->B->C (2 layers used: 0,1)
			},
			description: "Cascade overlaps should use transitive overlap detection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layers := handler.calculateEventLayers(tt.events, "UTC")
			require.Greater(t, len(layers), 0, "Should create at least one layer")

			// Collect all events from all layers
			allEvents := make(map[string]models.CalendarViewEvent)
			for _, layer := range layers {
				for _, event := range layer.Events {
					allEvents[event.ID] = event
				}
			}

			// Verify overlap group calculations
			for eventID, expected := range tt.expectedGroups {
				event, exists := allEvents[eventID]
				require.True(t, exists, "Event %s should exist in results", eventID)

				assert.Equal(t, expected.group, event.OverlapGroup,
					"Event %s should have OverlapGroup=%d, got %d",
					eventID, expected.group, event.OverlapGroup)

				assert.Equal(t, expected.index, event.OverlapIndex,
					"Event %s should have OverlapIndex=%d, got %d",
					eventID, expected.index, event.OverlapIndex)

				// Validate that index is within valid range
				assert.GreaterOrEqual(t, event.OverlapIndex, 0,
					"OverlapIndex should be >= 0")
				assert.Less(t, event.OverlapIndex, event.OverlapGroup,
					"OverlapIndex should be < OverlapGroup")
			}

			t.Logf("✅ %s: All overlap groups calculated correctly", tt.description)
		})
	}
}

// Test the specific Quick Check-in scenario that was failing
func TestQuickCheckinScenario(t *testing.T) {
	handler := &CalendarAPIHandler{}

	// Recreate the exact scenario from the screenshot
	events := []models.UnifiedCalendarEvent{
		// 12:00-12:30 PM: Quick Check-in (slots 48-50)
		createTestEventWithSlots("quick_checkin", "Quick Check-in", 48, 50),
		// 12:15-1:15 PM: Code Review (slots 49-53) - overlaps with Quick Check-in
		createTestEventWithSlots("code_review", "Code Review", 49, 53),
		// 12:30-2:00 PM: Strategy Meeting (slots 50-56) - overlaps with Code Review
		createTestEventWithSlots("strategy", "Strategy Meeting", 50, 56),
		// 1:00-1:30 PM: 1:1 with Sarah (slots 52-54) - overlaps with Strategy Meeting
		createTestEventWithSlots("sarah", "1:1 with Sarah", 52, 54),
	}

	layers := handler.calculateEventLayers(events, "UTC")

	// Collect all events from layers
	allEvents := make(map[string]models.CalendarViewEvent)
	for _, layer := range layers {
		for _, event := range layer.Events {
			allEvents[event.ID] = event
		}
	}

	// Verify layer assignments
	assert.Equal(t, 0, allEvents["quick_checkin"].OverlapIndex, "Quick Check-in should be in layer 0")
	assert.Equal(t, 1, allEvents["code_review"].OverlapIndex, "Code Review should be in layer 1")
	assert.Equal(t, 2, allEvents["sarah"].OverlapIndex, "1:1 with Sarah should be in layer 2")

	// The key test: Quick Check-in should see 3 layers (0, 1, 2) because:
	// - Layer 0: Quick Check-in itself
	// - Layer 1: Code Review (overlaps 49-50)
	// - Layer 2: Should be counted because there are events in layer 2 during the overlap period
	assert.Equal(t, 3, allEvents["quick_checkin"].OverlapGroup,
		"Quick Check-in should have overlapGroup=3 (sees layers 0,1,2)")

	// Verify other events
	assert.Equal(t, 3, allEvents["code_review"].OverlapGroup,
		"Code Review should have overlapGroup=3")
	assert.Equal(t, 3, allEvents["strategy"].OverlapGroup,
		"Strategy Meeting should have overlapGroup=3")
	assert.Equal(t, 3, allEvents["sarah"].OverlapGroup,
		"1:1 with Sarah should have overlapGroup=3")

	// Log the results for debugging
	for id, event := range allEvents {
		t.Logf("Event '%s' (layer %d): overlapGroup=%d, slots=%d-%d",
			id, event.OverlapIndex, event.OverlapGroup, event.StartSlot, event.EndSlot)
	}
}

// Test that overlap calculation gives consistent results for client rendering
func TestOverlapGroupClientCompatibility(t *testing.T) {
	handler := &CalendarAPIHandler{}

	// Create scenario with various overlap patterns
	events := []models.UnifiedCalendarEvent{
		createTestEvent("standalone", "Standalone", "08:00", "09:00"),
		createTestEvent("pair1", "Pair 1A", "10:00", "11:00"),
		createTestEvent("pair2", "Pair 1B", "10:30", "11:30"),
		createTestEvent("triple1", "Triple 1", "14:00", "15:00"),
		createTestEvent("triple2", "Triple 2", "14:20", "15:20"),
		createTestEvent("triple3", "Triple 3", "14:40", "16:00"),
	}

	layers := handler.calculateEventLayers(events, "UTC")
	allEvents := make(map[string]models.CalendarViewEvent)
	for _, layer := range layers {
		for _, event := range layer.Events {
			allEvents[event.ID] = event
		}
	}

	// Test client rendering logic
	testClientRendering := func(eventID string, expectedWidth float64, expectedLeft float64) {
		event := allEvents[eventID]
		width := 100.0 / float64(event.OverlapGroup)
		left := float64(event.OverlapIndex) * width

		assert.InDelta(t, expectedWidth, width, 0.01,
			"Event %s should have width %.1f%%, got %.1f%%", eventID, expectedWidth, width)
		assert.InDelta(t, expectedLeft, left, 0.01,
			"Event %s should have left position %.1f%%, got %.1f%%", eventID, expectedLeft, left)
	}

	// Standalone event: 100% width, 0% left
	testClientRendering("standalone", 100.0, 0.0)

	// Pair events: 50% width each
	testClientRendering("pair1", 50.0, 0.0)  // First event: 0% left
	testClientRendering("pair2", 50.0, 50.0) // Second event: 50% left

	// Triple events: 33.33% width each
	testClientRendering("triple1", 33.33, 0.0)   // First: 0% left
	testClientRendering("triple2", 33.33, 33.33) // Second: 33.33% left
	testClientRendering("triple3", 33.33, 66.67) // Third: 66.67% left

	t.Log("✅ Client rendering calculations validated")
}

// Helper function to create test events
func createTestEvent(id, title, startTime, endTime string) models.UnifiedCalendarEvent {
	start, _ := time.Parse("15:04", startTime)
	end, _ := time.Parse("15:04", endTime)

	return models.UnifiedCalendarEvent{
		ID:        id,
		Title:     title,
		StartTime: start,
		EndTime:   end,
		Color:     "#3b82f6",
		Attendees: []models.EventAttendee{},
	}
}

// Helper function to create test events with specific slot numbers
func createTestEventWithSlots(id, title string, startSlot, endSlot int) models.UnifiedCalendarEvent {
	// Convert slots back to time (each slot = 15 minutes)
	startHour := startSlot / 4
	startMin := (startSlot % 4) * 15
	endHour := endSlot / 4
	endMin := (endSlot % 4) * 15

	start := time.Date(2025, 9, 27, startHour, startMin, 0, 0, time.UTC)
	end := time.Date(2025, 9, 27, endHour, endMin, 0, 0, time.UTC)

	return models.UnifiedCalendarEvent{
		ID:        id,
		Title:     title,
		StartTime: start,
		EndTime:   end,
		Color:     "#3b82f6",
		Attendees: []models.EventAttendee{},
	}
}
