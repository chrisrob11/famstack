import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';

@customElement('daily-calendar')
export class DailyCalendar extends LitElement {
  @property({ type: String })
  date = new Date().toISOString().split('T')[0];

  @property({ type: Boolean, attribute: 'use-24hour' })
  use24Hour = false;

  static override styles = css`
    :host {
      --calendar-bg: #f8f9fa;
      --calendar-text: #333;
      --grid-border: #e1e5e9;
      --hour-text: #6c757d;
      --time-slot-height: 15px;
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

  private formatDate(dateString: string | undefined): string {
    if (!dateString) {
      dateString = new Date().toISOString().split('T')[0];
    }
    const date = new Date(dateString!);
    return date.toLocaleDateString('en-US', {
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
          position: ((hour - startHour) * 4 + quarter) * 15,
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

  private renderTimeGrid() {
    const slots = this.generateTimeSlots();
    return html`
      <div class="grid-content">
        ${slots.map(
          slot => html`
            <div class="time-row ${slot.isHourBoundary ? 'hour-boundary' : ''}">
              <div class="time-label-cell">${slot.isHourBoundary ? slot.timeString : ''}</div>
              <div class="time-slot" data-time="${slot.timeString}"></div>
            </div>
          `
        )}
      </div>
    `;
  }

  override firstUpdated() {
    // Scroll to 7 AM on initial load (7 hours from midnight start = 7 * 4 slots * 15px)
    const gridContent = this.shadowRoot?.querySelector('.grid-content') as HTMLElement;
    if (gridContent) {
      const scrollTo7AM = 7 * 4 * 15; // 7 hours * 4 quarters * 15px per slot
      gridContent.scrollTop = scrollTo7AM;
    }
  }

  private renderAllDaySection() {
    return html`
      <div class="all-day-section">
        <div class="all-day-label">All Day</div>
        <div class="all-day-events">
          <span class="all-day-placeholder">No all-day events</span>
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
