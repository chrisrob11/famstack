import { LitElement, html, css } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { styleMap } from 'lit/directives/style-map.js';
import { calendarApiService, UnifiedCalendarEvent, GetEventsOptions } from './services/calendar-api.js';
import './components/event-card.js';

const TIME_SLOT_HEIGHT_PX = 15;
const PIXEL_PER_MINUTE = TIME_SLOT_HEIGHT_PX / 15;

@customElement('daily-calendar')
export class DailyCalendar extends LitElement {
  @property({ type: String })
  date = new Date().toISOString().split('T')[0];

  @property({ type: Boolean, attribute: 'use-24hour' })
  use24Hour = false;

  @state()
  private _events: UnifiedCalendarEvent[] = [];

  @state()
  private _allDayEvents: UnifiedCalendarEvent[] = [];

  @state()
  private _timedEvents: UnifiedCalendarEvent[] = [];

  static override styles = css`
    :host {
      --calendar-bg: #f8f9fa;
      --calendar-text: #333;
      --grid-border: #e1e5e9;
      --hour-text: #6c757d;
      --time-slot-height: ${TIME_SLOT_HEIGHT_PX}px;
      --hour-line-color: #dee2e6;

      display: block;
      width: 100%;
      height: 100vh;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      background: var(--calendar-bg);
      color: var(--calendar-text);
    }

    .calendar-container {
      display: flex;
      flex-direction: column;
      height: 100%;
      padding: 16px;
      box-sizing: border-box;
    }

    .header {
      display: flex;
      justify-content: center;
      align-items: center;
      padding: 20px 0;
      border-bottom: 2px solid var(--grid-border);
      margin-bottom: 0;
      background: white;
      border-radius: 8px 8px 0 0;
    }

    .date-title {
      font-size: 24px;
      font-weight: 600;
      color: #2c3e50;
    }

    .all-day-section {
      background: white;
      border-bottom: 2px solid var(--grid-border);
      min-height: 60px;
      padding: 12px 16px;
      display: flex;
      align-items: center;
      font-size: 14px;
      color: var(--hour-text);
    }

    .all-day-label {
      width: 20px;
      font-weight: 600;
      text-align: right;
      padding-right: 2px;
      border-right: 2px solid var(--grid-border);
      margin-right: 16px;
    }

    .all-day-events {
      flex: 1;
      display: flex;
      gap: 8px;
      flex-wrap: wrap;
      align-items: center;
    }

    .all-day-placeholder {
      color: #adb5bd;
      font-style: italic;
    }

    .time-grid-container {
      flex: 1;
      background: white;
      border-radius: 0 0 8px 8px;
      overflow: hidden;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
      position: relative;
    }

    .grid-content {
      position: relative;
      overflow-y: auto;
      height: 100%;
    }

    .time-row {
      display: flex;
      height: var(--time-slot-height);
      border-bottom: 1px solid #f1f3f5;
    }

    .time-row.hour-boundary {
      border-bottom: 2px solid var(--hour-line-color);
    }

    .time-row:last-child {
      border-bottom: none;
    }

    .time-label-cell {
      width: 40px;
      background: #f8f9fa;
      border-right: 2px solid var(--grid-border);
      padding-right: 2px;
      padding-left: 7px;
      text-align: right;
      font-size: 12px;
      font-weight: 500;
      color: var(--hour-text);
      display: flex;
      align-items: flex-start;
      padding-top: 2px;
    }

    .time-slot {
      flex: 1;
      position: relative;
    }

    .events-container {
      position: absolute;
      top: 0;
      left: 42px; /* Width of label cell + border */
      right: 0;
      bottom: 0;
      z-index: 1;
    }

    .event-wrapper {
      position: absolute;
      left: 2px;
      right: 8px; /* Gap on the right side */
      overflow: hidden;
    }

    .grid-content::-webkit-scrollbar {
      width: 8px;
    }

    .grid-content::-webkit-scrollbar-track {
      background: #f1f1f1;
    }

    .grid-content::-webkit-scrollbar-thumb {
      background: #c1c1c1;
      border-radius: 4px;
    }

    .grid-content::-webkit-scrollbar-thumb:hover {
      background: #a8a8a8;
    }

    @media (max-width: 1024px) {
      .calendar-container {
        padding: 8px;
      }

      .header {
        padding: 16px 0;
      }

      .date-title {
        font-size: 20px;
      }
    }
  `;

  private async _fetchEvents() {
    const options: GetEventsOptions = {};
    if (this.date) {
      options.date = this.date;
    }
    this._events = await calendarApiService.getUnifiedCalendarEvents(options);
    this._processEvents();
  }

  private _processEvents() {
    this._allDayEvents = this._events.filter(e => e.all_day);
    this._timedEvents = this._events.filter(e => !e.all_day);
  }

  private formatDate(dateString: string | undefined): string {
    if (!dateString) {
      dateString = new Date().toISOString().split('T')[0];
    }
    const date = new Date(dateString!);
    return date.toLocaleDateString(undefined, {
      weekday: 'long',
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  }

  private generateTimeSlots() {
    const slots = [];
    const startHour = 0; // Start at midnight (12 AM)
    const endHour = 24;

    for (let hour = startHour; hour < endHour; hour++) {
      for (let quarter = 0; quarter < 4; quarter++) {
        const minutes = quarter * 15;
        const isHourBoundary = quarter === 0;
        const timeString = this.formatTimeSlot(hour, minutes);

        slots.push({
          hour,
          minutes,
          timeString,
          isHourBoundary,
          position: ((hour - startHour) * 4 + quarter) * TIME_SLOT_HEIGHT_PX,
        });
      }
    }

    return slots;
  }

  private formatTimeSlot(hour: number, minutes: number): string {
    if (this.use24Hour) {
      // 24-hour format
      const hourStr = hour.toString().padStart(2, '0');
      return minutes === 0 ? `${hourStr}:00` : `${hourStr}:${minutes.toString().padStart(2, '0')}`;
    } else {
      // 12-hour format
      const period = hour >= 12 ? 'PM' : 'AM';
      const displayHour = hour > 12 ? hour - 12 : hour === 0 ? 12 : hour;
      return minutes === 0
        ? `${displayHour} ${period}`
        : `${displayHour}:${minutes.toString().padStart(2, '0')} ${period}`;
    }
  }

  private renderTimedEvents() {
    return html`
      ${this._timedEvents.map(event => {
        const startTime = new Date(event.start_time);
        const endTime = new Date(event.end_time);

        const startMinutes = startTime.getHours() * 60 + startTime.getMinutes();
        const durationMinutes = (endTime.getTime() - startTime.getTime()) / 60000;

        const top = startMinutes * PIXEL_PER_MINUTE;
        const height = durationMinutes * PIXEL_PER_MINUTE - 2; // -2 for a small gap

        const eventStyles = {
          top: `${top}px`,
          height: `${height}px`,
        };

        return html`
          <div class="event-wrapper" style=${styleMap(eventStyles)}>
            <event-card .event=${event}></event-card>
          </div>
        `;
      })}
    `;
  }

  private renderTimeGrid() {
    const slots = this.generateTimeSlots();
    return html`
      <div class="grid-content" role="grid" aria-label="Daily calendar time grid">
        ${slots.map(
          slot => html`
            <div class="time-row ${slot.isHourBoundary ? 'hour-boundary' : ''}" role="row">
              <div class="time-label-cell" role="rowheader">
                ${slot.isHourBoundary ? slot.timeString : ''}
              </div>
              <div class="time-slot" role="gridcell" aria-label=${slot.timeString}></div>
            </div>
          `
        )}
        <div class="events-container">
          ${this.renderTimedEvents()}
        </div>
      </div>
    `;
  }

  override firstUpdated() {
    this._fetchEvents();
    // Scroll to 7 AM on initial load
    const gridContent = this.shadowRoot?.querySelector('.grid-content') as HTMLElement;
    if (gridContent) {
      const scrollTo7AM = 7 * 4 * TIME_SLOT_HEIGHT_PX; // 7 hours * 4 quarters * height
      gridContent.scrollTop = scrollTo7AM;
    }
  }

  private renderAllDaySection() {
    return html`
      <div class="all-day-section" role="region" aria-labelledby="all-day-heading">
        <div id="all-day-heading" class="all-day-label" role="rowheader">All Day</div>
        <div class="all-day-events" role="grid">
          ${this._allDayEvents.length > 0
            ? this._allDayEvents.map(
                event => html`<event-card .event=${event}></event-card>`
              )
            : html`<span class="all-day-placeholder">No all-day events</span>`}
        </div>
      </div>
    `;
  }

  override render() {
    return html`
      <div class="calendar-container">
        <div class="header">
          <h1 class="date-title">${this.formatDate(this.date)}</h1>
        </div>
        ${this.renderAllDaySection()}
        <div class="time-grid-container">${this.renderTimeGrid()}</div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'daily-calendar': DailyCalendar;
  }
}
