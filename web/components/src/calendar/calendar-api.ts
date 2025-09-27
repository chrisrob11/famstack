// Event attendee with display information for person identification
export interface EventAttendee {
  id: string;
  name: string;
  initial: string;
  color: string;
  response: string; // needsAction, accepted, declined, tentative
}

// Corresponds to the Go model: internal/models/calendar.go
export interface UnifiedCalendarEvent {
  id: string;
  family_id: string;
  title: string;
  description?: string;
  start_time: string; // ISO 8601 format
  end_time: string; // ISO 8601 format
  all_day: boolean;
  event_type: string;
  color: string;
  created_by?: string;
  priority: number;
  status: string;
  created_at: string;
  updated_at: string;
  attendees: EventAttendee[]; // Array of attendees with full display data
}

export interface GetEventsOptions {
  date?: string;
  startDate?: string;
  endDate?: string;
  familyId?: string;
}

// New layered calendar data structures
export interface CalendarViewEvent {
  id: string;
  title: string;
  startSlot: number; // 0-359 (15-minute intervals)
  endSlot: number; // 0-359
  color: string;
  ownerId: string;
  attendeeIds: string[];
  overlapGroup: number; // Total events in this overlap group
  overlapIndex: number; // Position within overlap group (0-based)
  attendees: EventAttendee[];
  isPrivate: boolean;
  location?: string;
  description?: string;
}

export interface CalendarLayer {
  layerIndex: number;
  events: CalendarViewEvent[];
}

export interface DayView {
  date: string; // YYYY-MM-DD
  layers: CalendarLayer[];
}

export interface DaysResponse {
  startDate: string;
  endDate: string;
  timezone: string;
  requestedPeople: string[];
  days: DayView[];
  metadata: {
    totalEvents: number;
    lastUpdated: string;
    maxDaysLimit: number;
  };
}

export interface GetDaysOptions {
  startDate: string; // Required: YYYY-MM-DD
  endDate: string; // Required: YYYY-MM-DD
  people?: string[]; // Optional: array of person IDs
  timezone?: string; // Optional: timezone string
}

class CalendarApiService {
  private readonly BASE_URL = '/api/v1/calendar';

  public async getUnifiedCalendarEvents(
    options: GetEventsOptions
  ): Promise<UnifiedCalendarEvent[]> {
    const url = new URL(`${this.BASE_URL}/events`, window.location.origin);

    if (options.date) {
      url.searchParams.append('date', options.date);
    }
    if (options.startDate) {
      url.searchParams.append('start_date', options.startDate);
    }
    if (options.endDate) {
      url.searchParams.append('end_date', options.endDate);
    }
    if (options.familyId) {
      url.searchParams.append('family_id', options.familyId);
    }

    try {
      const response = await fetch(url.toString());
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const events = await response.json();
      // The backend ensures we always get an array
      return events as UnifiedCalendarEvent[];
    } catch (error) {
      // Throw error instead of silently returning empty array
      throw new Error(
        `Failed to fetch calendar events: ${error instanceof Error ? error.message : 'Unknown error'}`
      );
    }
  }

  /**
   * Get layered calendar days with pre-calculated event positioning
   */
  public async getCalendarDays(options: GetDaysOptions): Promise<DaysResponse> {
    const url = new URL(`${this.BASE_URL}/days`, window.location.origin);

    // Required parameters
    url.searchParams.append('startDate', options.startDate);
    url.searchParams.append('endDate', options.endDate);

    // Optional parameters
    if (options.people && options.people.length > 0) {
      url.searchParams.append('people', options.people.join(','));
    }
    if (options.timezone) {
      url.searchParams.append('timezone', options.timezone);
    }

    try {
      const response = await fetch(url.toString());
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const daysData = await response.json();
      return daysData as DaysResponse;
    } catch (error) {
      throw new Error(
        `Failed to fetch calendar days: ${error instanceof Error ? error.message : 'Unknown error'}`
      );
    }
  }
}

// Export a singleton instance of the service
export const calendarApiService = new CalendarApiService();
