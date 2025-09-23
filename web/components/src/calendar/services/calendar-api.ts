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
  end_time: string;   // ISO 8601 format
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
      console.error('❌ Failed to fetch calendar events:', error);
      // Return an empty array on error to prevent frontend crashes
      return [];
    }
  }
}

// Export a singleton instance of the service
export const calendarApiService = new CalendarApiService();
