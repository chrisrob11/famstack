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
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
      cursor: pointer;
      transition: background-color 0.2s;
    }

    :host(:hover) {
      opacity: 0.9;
    }

    .title {
      font-weight: 500;
    }
  `;

  override render() {
    if (!this.event) {
      return html``;
    }

    const cardStyles = {
      backgroundColor: this.event.color || '#3b82f6', // Default color
    };

    return html`
      <div style=${styleMap(cardStyles)}>
        <span class="title">${this.event.title}</span>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'event-card': EventCard;
  }
}
