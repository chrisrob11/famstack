import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';

@customElement('daily-calendar')
export class DailyCalendar extends LitElement {
  @property({ type: String })
  date = new Date().toISOString().split('T')[0];

  static override styles = css`
    :host {
      display: block;
      width: 100%;
      height: 100vh;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      background: #f8f9fa;
      color: #333;
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
      border-bottom: 1px solid #e1e5e9;
      margin-bottom: 20px;
    }

    .date-title {
      font-size: 24px;
      font-weight: 600;
      color: #2c3e50;
    }

    .hello-world {
      display: flex;
      justify-content: center;
      align-items: center;
      flex: 1;
      font-size: 32px;
      font-weight: 300;
      color: #7f8c8d;
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
      day: 'numeric'
    });
  }

  override render() {
    return html`
      <div class="calendar-container">
        <div class="header">
          <h1 class="date-title">${this.formatDate(this.date)}</h1>
        </div>
        <div class="hello-world">
          🎉 Hello World - Daily Calendar Component 🎉
          <br><br>
          <small>✅ Milestone 1 Complete: Foundation Setup with Lit Framework</small>
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'daily-calendar': DailyCalendar;
  }
}