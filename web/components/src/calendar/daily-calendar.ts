import { LitElement, html, css } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { TimeFormatter } from './utils/time-formatter.js';
import { CALENDAR_CONFIG } from './calendar-config.js';
import { calendarApiService, type DayView, type CalendarViewEvent } from './calendar-api.js';
import { styleMap } from 'lit/directives/style-map.js';

// import './event-card.js'; // Not needed in layered approach

@customElement('daily-calendar')
export class DailyCalendar extends LitElement {
  constructor() {
    super();
    this._timeFormatter = new TimeFormatter(this.use24Hour);
  }

  @property({ type: String })
  date = new Date().toISOString().split('T')[0];

  @property({ type: Boolean, attribute: 'use-24hour' })
  use24Hour = false;

  @property({ type: Array })
  people: string[] = [];

  @state()
  private _dayView: DayView | null = null;

  @state()
  private _isLoading = false;

  @state()
  private _errorState: string | null = null;

  private _timeFormatter: TimeFormatter;

  static override styles = css`
    :host {
      --calendar-bg: #f8f9fa;
      --calendar-text: #333;
      --grid-border: #e1e5e9;
      --hour-text: #6c757d;
      --time-slot-height: ${CALENDAR_CONFIG.TIME_SLOT_HEIGHT}px;
      --hour-line-color: #dee2e6;

      display: block;
      width: 100%;
      height: 100vh;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      background: var(--calendar-bg);
      color: var(--calendar-text);
      overflow: hidden;
    }

    .calendar-container {
      height: 100%;
      display: flex;
      flex-direction: column;
    }

    .time-grid {
      flex: 1;
      display: grid;
      grid-template-rows: repeat(96, var(--time-slot-height)); /* 96 * 15min = 24 hours */
      position: relative;
      overflow-y: auto;
      border-left: 1px solid var(--grid-border);
    }

    .time-labels {
      position: absolute;
      left: 0;
      top: 0;
      width: 60px;
      height: 100%;
      background: var(--calendar-bg);
      border-right: 1px solid var(--grid-border);
      z-index: 10;
    }

    .hour-label {
      position: absolute;
      left: 8px;
      transform: translateY(-50%);
      font-size: 12px;
      color: var(--hour-text);
      font-weight: 500;
    }

    .hour-line {
      position: absolute;
      left: 60px;
      right: 0;
      height: 1px;
      background: var(--hour-line-color);
      z-index: 5;
    }

    .events-container {
      position: absolute;
      left: 60px;
      right: 8px;
      top: 0;
      height: 100%;
    }

    .calendar-event {
      position: absolute;
      pointer-events: auto;
      border-radius: 4px;
      padding: 4px 8px;
      font-size: 12px;
      font-weight: 500;
      cursor: pointer;
      border: 1px solid transparent;
      transition: all 0.2s ease;
      overflow: hidden;
      box-sizing: border-box;
      min-height: 20px;
      margin-bottom: 3px; /* Add gap between adjacent events */
    }

    .calendar-event:hover {
      border-color: rgba(255, 255, 255, 0.9);
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
      z-index: 100;
      transform: scale(1.02);
    }

    .calendar-event:focus {
      outline: none;
      border-color: #007bff;
      box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.25);
      z-index: 101;
    }

    .loading-spinner {
      display: flex;
      justify-content: center;
      align-items: center;
      height: 200px;
      font-size: 14px;
      color: var(--hour-text);
    }

    .error-message {
      padding: 16px;
      background: #fee;
      color: #c00;
      border-radius: 4px;
      margin: 16px;
    }

    .event-content {
      position: relative;
      height: 100%;
      width: 100%;
    }

    .event-title {
      position: absolute;
      left: 4px;
      right: 4px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      line-height: 1.2;
    }

    .event-title.short-event {
      top: 50%;
      transform: translateY(-50%);
    }

    .event-title.long-event {
      top: 4px;
    }

    .event-attendees {
      position: absolute;
      bottom: 4px;
      right: 4px;
      display: flex;
      gap: 2px;
      flex-shrink: 0;
    }

    .event-attendees.centered {
      top: 50%;
      bottom: auto;
      transform: translateY(-50%);
    }

    .attendee-avatar {
      width: 16px;
      height: 16px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 10px;
      font-weight: 600;
      color: white;
      border: 1px solid rgba(255, 255, 255, 0.3);
    }

    .attendee-avatar.small {
      width: 12px;
      height: 12px;
      font-size: 8px;
    }
  `;

  override connectedCallback() {
    super.connectedCallback();
    this._loadCalendarData();
  }

  override updated(changedProperties: Map<string, any>) {
    if (changedProperties.has('date') || changedProperties.has('people')) {
      this._loadCalendarData();
    }
  }

  private async _loadCalendarData() {
    this._isLoading = true;
    this._errorState = null;

    try {
      const options: any = {
        startDate: this.date,
        endDate: this.date,
      };

      if (this.people.length > 0) {
        options.people = this.people;
      }

      const response = await calendarApiService.getCalendarDays(options);

      this._dayView = response.days.find(day => day.date === this.date) || null;
    } catch (error) {
      this._errorState = error instanceof Error ? error.message : 'Failed to load calendar data';
      this._dayView = null;
    } finally {
      this._isLoading = false;
    }
  }

  private _renderTimeGrid() {
    const hours = Array.from({ length: 24 }, (_, i) => i);

    return html`
      <div class="time-labels">
        ${hours.map(
          hour => html`
            <div class="hour-label" style="top: ${hour * 4 * CALENDAR_CONFIG.TIME_SLOT_HEIGHT}px">
              ${this._timeFormatter.formatTimeSlot(hour, 0)}
            </div>
          `
        )}
      </div>

      ${hours.map(
        hour => html`
          <div
            class="hour-line"
            style="top: ${hour * 4 * CALENDAR_CONFIG.TIME_SLOT_HEIGHT}px"
          ></div>
        `
      )}
    `;
  }

  private _renderEventLayers() {
    if (!this._dayView || this._dayView.layers.length === 0) {
      return html``;
    }

    // Collect all events from all layers into a single list
    const allEvents: CalendarViewEvent[] = [];
    for (const layer of this._dayView.layers) {
      allEvents.push(...layer.events);
    }

    return html`
      <div class="events-container">${allEvents.map(event => this._renderEvent(event))}</div>
    `;
  }

  private _renderEvent(event: CalendarViewEvent) {
    const topPx = event.startSlot * CALENDAR_CONFIG.TIME_SLOT_HEIGHT;
    // Subtract 3px from height to account for margin-bottom gap
    const heightPx = (event.endSlot - event.startSlot) * CALENDAR_CONFIG.TIME_SLOT_HEIGHT - 3;

    // Check if event is 15 minutes or less (1 slot = 15 minutes)
    const isShortEvent = event.endSlot - event.startSlot <= 1;
    const titleClass = isShortEvent ? 'event-title short-event' : 'event-title long-event';

    // Calculate width and position based on overlap group
    const widthPercent = 100 / event.overlapGroup;
    const leftPercent = event.overlapIndex * widthPercent;

    return html`
      <div
        class="calendar-event"
        style=${styleMap({
          top: `${topPx}px`,
          height: `${heightPx}px`,
          width: `${widthPercent}%`,
          left: `${leftPercent}%`,
          backgroundColor: event.color,
          color: this._getContrastColor(event.color),
        })}
        tabindex="0"
        @click=${() => this._handleEventClick(event)}
        @keydown=${(e: KeyboardEvent) => this._handleEventKeydown(e, event)}
      >
        <div class="event-content">
          <div class="${titleClass}">${event.title}</div>
          ${event.attendees && event.attendees.length > 0
            ? html`
                <div class="event-attendees ${isShortEvent ? 'centered' : ''}">
                  ${event.attendees.slice(0, 3).map(
                    attendee => html`
                      <div
                        class="attendee-avatar ${isShortEvent ? 'small' : ''}"
                        style=${styleMap({
                          backgroundColor: attendee.color,
                        })}
                        title="${attendee.name}"
                      >
                        ${attendee.initial}
                      </div>
                    `
                  )}
                  ${event.attendees.length > 3
                    ? html`
                        <div
                          class="attendee-avatar ${isShortEvent ? 'small' : ''}"
                          style="background-color: #666;"
                          title="${event.attendees.length - 3} more attendees"
                        >
                          +${event.attendees.length - 3}
                        </div>
                      `
                    : ''}
                </div>
              `
            : ''}
        </div>
      </div>
    `;
  }

  private _getContrastColor(backgroundColor: string): string {
    // Simple contrast calculation - in a real app you'd want something more sophisticated
    const color = backgroundColor.replace('#', '');
    const r = parseInt(color.substring(0, 2), 16);
    const g = parseInt(color.substring(2, 4), 16);
    const b = parseInt(color.substring(4, 6), 16);
    const brightness = (r * 299 + g * 587 + b * 114) / 1000;
    return brightness > 128 ? '#000' : '#fff';
  }

  private _handleEventClick(event: CalendarViewEvent) {
    // Dispatch custom event for parent components to handle
    this.dispatchEvent(
      new CustomEvent('event-click', {
        detail: { event },
        bubbles: true,
      })
    );
  }

  private _handleEventKeydown(e: KeyboardEvent, event: CalendarViewEvent) {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      this._handleEventClick(event);
    }
  }

  override render() {
    if (this._isLoading) {
      return html` <div class="loading-spinner">Loading calendar...</div> `;
    }

    if (this._errorState) {
      return html` <div class="error-message">Error: ${this._errorState}</div> `;
    }

    return html`
      <div class="calendar-container">
        <div class="time-grid">${this._renderTimeGrid()} ${this._renderEventLayers()}</div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'daily-calendar': DailyCalendar;
  }
}
