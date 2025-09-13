import { ComponentConfig } from '../common/types.js';

export interface CalendarEvent {
  id: string;
  title: string;
  start_time: string;
  end_time: string;
  description?: string;
  attendees?: string[];
}

/**
 * Calendar service for fetching calendar events
 */
export class CalendarService {
  private config: ComponentConfig;

  constructor(config: ComponentConfig) {
    this.config = config;
  }

  async getEventsForDate(date: string): Promise<CalendarEvent[]> {
    const response = await fetch(`${this.config.apiBaseUrl}/calendar/events?date=${date}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': this.config.csrfToken || '',
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch calendar events: ${response.statusText}`);
    }

    return response.json();
  }

  async getEventsForDateRange(startDate: string, endDate: string): Promise<CalendarEvent[]> {
    const response = await fetch(
      `${this.config.apiBaseUrl}/calendar/events?start_date=${startDate}&end_date=${endDate}`,
      {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': this.config.csrfToken || '',
        },
      }
    );

    if (!response.ok) {
      throw new Error(`Failed to fetch calendar events: ${response.statusText}`);
    }

    return response.json();
  }

  async createEvent(eventData: Partial<CalendarEvent>): Promise<CalendarEvent> {
    const response = await fetch(`${this.config.apiBaseUrl}/calendar/events`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': this.config.csrfToken || '',
      },
      body: JSON.stringify(eventData),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText || `Failed to create calendar event: ${response.statusText}`);
    }

    return response.json();
  }

  async updateEvent(eventId: string, updates: Partial<CalendarEvent>): Promise<CalendarEvent> {
    const response = await fetch(`${this.config.apiBaseUrl}/calendar/events/${eventId}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': this.config.csrfToken || '',
      },
      body: JSON.stringify(updates),
    });

    if (!response.ok) {
      throw new Error(`Failed to update calendar event: ${response.statusText}`);
    }

    return response.json();
  }

  async deleteEvent(eventId: string): Promise<void> {
    const response = await fetch(`${this.config.apiBaseUrl}/calendar/events/${eventId}`, {
      method: 'DELETE',
      headers: {
        'X-CSRF-Token': this.config.csrfToken || '',
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to delete calendar event: ${response.statusText}`);
    }
  }
}
