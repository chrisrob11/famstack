import { ComponentConfig } from '../common/types.js';
import { CalendarService, CalendarEvent } from './calendar-events-service.js';
import { CalendarEventComponent, EventData } from './calendar-event.js';

/**
 * Main CalendarEvents component that displays events in a vertical timeline format
 * Integrates with the existing calendar service to get unified calendar events
 */
export class CalendarEvents {
  private container: HTMLElement;
  private config: ComponentConfig;
  private calendarService: CalendarService;
  private events: CalendarEvent[] = [];
  private currentDate: Date = new Date();
  private timelineStartHour: number = 6; // 6 AM
  private timelineEndHour: number = 22; // 10 PM

  // Default person colors for attendee circles
  private personColors: { [key: string]: string } = {
    user1: '#3b82f6', // Blue for Dad
    user2: '#8b5cf6', // Purple for Mom
    user3: '#10b981', // Green for Alex/Child
    Dad: '#3b82f6',
    Mom: '#8b5cf6',
    Alex: '#10b981',
    user_dad: '#3b82f6',
    user_mom: '#8b5cf6',
    user_alex: '#10b981',
  };

  constructor(container: HTMLElement, config: ComponentConfig, calendarService: CalendarService) {
    this.container = container;
    this.config = config;
    this.calendarService = calendarService;
    this.init();
  }

  private init(): void {
    this.loadEvents();
  }

  private async loadEvents(): Promise<void> {
    try {
      this.renderLoading();
      const dateStr = this.currentDate.toISOString().split('T')[0]!;
      this.events = await this.calendarService.getEventsForDate(dateStr);
      this.renderCalendarEvents();
    } catch (error) {
      this.renderError('Failed to load events');
    }
  }

  private renderLoading(): void {
    this.container.innerHTML = `
      <div class="calendar-events-loading">
        <div class="loading-spinner"></div>
        <p>Loading events...</p>
      </div>
    `;
  }

  private renderError(message: string): void {
    this.container.innerHTML = `
      <div class="calendar-events-error">
        <p class="error-message">${message}</p>
        <button class="retry-btn" data-action="retry">Try Again</button>
      </div>
    `;
  }

  private renderCalendarEvents(): void {
    this.container.innerHTML = `
      <div class="calendar-events">
        <div class="calendar-events-container">
          ${this.renderTimelineGrid()}
        </div>
      </div>
    `;

    // Set up event listeners
    this.setupEventListeners();

    // Render events
    this.renderEvents();
  }

  private renderTimelineGrid(): string {
    const totalHours = this.timelineEndHour - this.timelineStartHour;
    const gridHeight = totalHours * 60; // 60px per hour

    let timeLabels = '';
    let gridLines = '';

    for (let hour = this.timelineStartHour; hour <= this.timelineEndHour; hour++) {
      const top = (hour - this.timelineStartHour) * 60;
      const timeLabel = this.formatTimeLabel(hour);

      timeLabels += `
        <div class="time-label" style="top: ${top}px;">
          ${timeLabel}
        </div>
      `;

      if (hour < this.timelineEndHour) {
        gridLines += `
          <div class="grid-line" style="top: ${top}px;"></div>
        `;
      }
    }

    return `
      <div class="calendar-events-grid" style="height: ${gridHeight}px; position: relative;">
        <div class="time-labels">
          ${timeLabels}
        </div>
        <div class="grid-lines">
          ${gridLines}
        </div>
        <div class="events-container" id="events-container" style="position: relative; height: 100%; margin-left: 60px;">
          <!-- Events will be rendered here -->
        </div>
      </div>
    `;
  }

  private formatTimeLabel(hour: number): string {
    if (hour === 0) return '12 AM';
    if (hour < 12) return `${hour} AM`;
    if (hour === 12) return '12 PM';
    return `${hour - 12} PM`;
  }

  private renderEvents(): void {
    const eventsContainer = this.container.querySelector('#events-container') as HTMLElement;
    if (!eventsContainer) return;

    // Clear existing events
    eventsContainer.innerHTML = '';

    // Convert calendar events to timeline events
    const timelineEvents = this.convertCalendarEventsToTimelineEvents(this.events);

    // Calculate columns for overlapping events
    const eventsWithColumns = this.calculateEventColumns(timelineEvents);

    // Render each event
    eventsWithColumns.forEach(eventData => {
      const eventComponent = new CalendarEventComponent(eventData, this.personColors);
      const timelineStartMinutes = this.timelineStartHour * 60;
      eventComponent.render(eventsContainer, timelineStartMinutes);
    });
  }

  private convertCalendarEventsToTimelineEvents(calendarEvents: CalendarEvent[]): EventData[] {
    return calendarEvents.map(event => {
      // Convert attendee IDs to attendee objects with names
      const attendees =
        event.attendees?.map((attendeeId: string) => ({
          name: this.getAttendeeNameById(attendeeId),
        })) || [];

      return {
        unifiedId: event.id,
        title: event.title,
        start: event.start_time,
        end: event.end_time || event.start_time,
        sources: ['calendar'],
        attendees: attendees,
        column: 0, // Will be calculated later
        totalColumns: 1, // Will be calculated later
      };
    });
  }

  private getAttendeeNameById(attendeeId: string): string {
    // Map user IDs to names - in a real app, this would come from user service
    const userNames: { [key: string]: string } = {
      user1: 'Dad',
      user2: 'Mom',
      user3: 'Alex',
      user_dad: 'Dad',
      user_mom: 'Mom',
      user_alex: 'Alex',
    };

    return userNames[attendeeId] || attendeeId.charAt(0).toUpperCase();
  }

  private calculateEventColumns(events: EventData[]): EventData[] {
    if (events.length === 0) return events;

    // Sort events by start time
    const sortedEvents = events.sort(
      (a, b) => new Date(a.start).getTime() - new Date(b.start).getTime()
    );

    // Group overlapping events
    const eventGroups: EventData[][] = [];

    for (const event of sortedEvents) {
      let placed = false;

      for (const group of eventGroups) {
        const hasOverlap = group.some(groupEvent => this.eventsOverlap(event, groupEvent));

        if (hasOverlap) {
          group.push(event);
          placed = true;
          break;
        }
      }

      if (!placed) {
        eventGroups.push([event]);
      }
    }

    // Assign columns within each group
    eventGroups.forEach(group => {
      const totalColumns = group.length;

      group.forEach((event, index) => {
        event.column = index;
        event.totalColumns = totalColumns;
      });
    });

    return sortedEvents;
  }

  private eventsOverlap(event1: EventData, event2: EventData): boolean {
    const start1 = new Date(event1.start);
    const end1 = new Date(event1.end);
    const start2 = new Date(event2.start);
    const end2 = new Date(event2.end);

    return start1 < end2 && start2 < end1;
  }

  private setupEventListeners(): void {
    const retryButton = this.container.querySelector('[data-action="retry"]');
    if (retryButton) {
      retryButton.addEventListener('click', () => {
        this.loadEvents();
      });
    }
  }

  public setDate(date: Date): void {
    this.currentDate = date;
    this.loadEvents();
  }

  public async refresh(): Promise<void> {
    await this.loadEvents();
  }

  public destroy(): void {
    // Clean up event listeners if needed
  }
}
