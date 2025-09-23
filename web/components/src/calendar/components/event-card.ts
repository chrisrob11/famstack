import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { styleMap } from 'lit/directives/style-map.js';
import type { UnifiedCalendarEvent } from '../services/calendar-api.js';

@customElement('event-card')
export class EventCard extends LitElement {
  @property({ type: Object })
  event?: UnifiedCalendarEvent;

  static override styles = css`
    :host {
      display: block;
      border-radius: 4px;
      padding: 4px 6px;
      font-size: 12px;
      color: white;
      box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
      cursor: pointer;
      transition: background-color 0.2s;
      height: 100%; /* Fill the wrapper height */
      box-sizing: border-box;
    }

    :host(:hover) {
      opacity: 0.9;
    }

    .event-content {
      display: flex;
      align-items: center;
      gap: 6px;
      height: 100%;
    }

    .title {
      font-weight: 500;
      flex: 1;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .attendees {
      display: flex;
      gap: 2px;
      flex-shrink: 0;
    }

    .person-circle {
      width: 16px;
      height: 16px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 8px;
      font-weight: 600;
      color: white;
      text-shadow: 0 0 2px rgba(0, 0, 0, 0.3);
      border: 1px solid rgba(255, 255, 255, 0.3);
    }
  `;

  override render() {
    if (!this.event) {
      return html``;
    }

    const cardStyles = {
      backgroundColor: this.event.color || '#3b82f6', // Default color
    };

    // Render attendees as person circles with initials
    const attendeeCircles = this.event.attendees?.map(attendee => {
      const circleStyles = {
        backgroundColor: attendee.color || '#6b7280'
      };

      return html`
        <div
          class="person-circle"
          style=${styleMap(circleStyles)}
          title="${attendee.name}"
        >
          ${attendee.initial}
        </div>
      `;
    }) || [];

    return html`
      <div style=${styleMap(cardStyles)}>
        <div class="event-content">
          <span class="title">${this.event.title}</span>
          ${attendeeCircles.length > 0 ? html`
            <div class="attendees">
              ${attendeeCircles}
            </div>
          ` : ''}
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'event-card': EventCard;
  }
}
