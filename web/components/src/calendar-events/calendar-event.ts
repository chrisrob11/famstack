export interface EventData {
  unifiedId: string;
  title: string;
  start: string;
  end: string;
  sources: string[];
  attendees: { name: string }[];
  column: number;
  totalColumns: number;
}

/**
 * CalendarEventComponent - Renders individual events in the calendar-events timeline
 * Originally provided by the user for timeline integration, now renamed for calendar events
 */
export class CalendarEventComponent {
  private event: EventData;
  private personColors: { [key: string]: string };

  constructor(event: EventData, personColors: { [key: string]: string }) {
    this.event = event;
    this.personColors = personColors;
  }

  public render(parentElement: HTMLElement, timelineStartMinutes: number): void {
    const start = new Date(this.event.start);
    const end = new Date(this.event.end);
    const durationMinutes = (end.getTime() - start.getTime()) / (1000 * 60);
    const startMinutes = start.getHours() * 60 + start.getMinutes();

    // Calculate position and size - 60px per hour
    const top = ((startMinutes - timelineStartMinutes) / 60) * 60 + 'px';
    const height = (durationMinutes / 60) * 60 + 'px';

    // Calculate width and left offset based on column
    const columnWidth = 100 / this.event.totalColumns;
    const left = this.event.column * columnWidth;
    const width = columnWidth - 2; // Leave some space between columns

    const eventElement = document.createElement('div');
    eventElement.className = 'calendar-event';
    eventElement.style.top = top;
    eventElement.style.height = height;
    eventElement.style.left = `${left}%`;
    eventElement.style.width = `${width}%`;
    eventElement.setAttribute('data-event-id', this.event.unifiedId);

    // Add event type data attribute for styling
    if (this.event.sources.length > 0) {
      eventElement.setAttribute('data-event-type', this.event.sources[0] || 'event');
    }

    // Format time for display
    const startTime = this.formatTime(start);
    const endTime = this.formatTime(end);
    const timeDisplay = `${startTime} - ${endTime}`;

    // Create attendee circles
    const attendeeCircles = this.event.attendees
      .map(attendee => {
        const initial = attendee.name.charAt(0).toUpperCase();
        const color =
          this.personColors[attendee.name] ||
          this.personColors[attendee.name.toLowerCase()] ||
          '#6b7280';

        return `<div class="attendee-circle" style="background-color: ${color};">${initial}</div>`;
      })
      .join('');

    eventElement.innerHTML = `
      <h3>${this.event.title}</h3>
      <div class="event-time">${timeDisplay}</div>
      ${attendeeCircles ? `<div class="event-attendees">${attendeeCircles}</div>` : ''}
    `;

    parentElement.appendChild(eventElement);
  }

  private formatTime(date: Date): string {
    const hours = date.getHours();
    const minutes = date.getMinutes();
    const ampm = hours >= 12 ? 'PM' : 'AM';
    const displayHours = hours % 12 || 12;
    const displayMinutes = minutes === 0 ? '' : `:${minutes.toString().padStart(2, '0')}`;
    return `${displayHours}${displayMinutes} ${ampm}`;
  }
}
