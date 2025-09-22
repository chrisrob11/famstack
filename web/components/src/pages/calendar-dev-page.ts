
import { BasePage } from './base-page.js';
import { ComponentConfig } from '../common/types.js';

export class CalendarDevPage extends BasePage {
  constructor(container: HTMLElement, config: ComponentConfig) {
    super(container, config, 'calendar-dev');
  }

  async init(): Promise<void> {
    this.render();
    this.loadCalendarComponent();
  }

  private render(): void {
    this.container.innerHTML = `
      <div class="calendar-dev-page">
        <div class="dev-header">
          <h1>📅 Daily Calendar Component - Development</h1>
          <p>Milestone 1: Foundation Setup - Basic component with "Hello World"</p>
          <div class="dev-status">
            <span class="status-badge">✅ Lit Framework Installed</span>
            <span class="status-badge">✅ Component Structure Created</span>
            <span class="status-badge">🔄 Testing Environment Active</span>
          </div>
        </div>
        <div class="component-container">
          <daily-calendar></daily-calendar>
        </div>
      </div>
    `;

    this.addStyles();
  }

  private addStyles(): void {
    const styles = `
      <style id="calendar-dev-page-styles">
        .calendar-dev-page {
          width: 100%;
          height: 100vh;
          display: flex;
          flex-direction: column;
          background: #f8f9fa;
        }

        .dev-header {
          background: #fff;
          padding: 1.5rem 2rem;
          border-bottom: 2px solid #e9ecef;
          box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }

        .dev-header h1 {
          font-size: 1.8rem;
          font-weight: 700;
          color: #2c3e50;
          margin: 0 0 0.5rem 0;
        }

        .dev-header p {
          color: #6c757d;
          font-size: 1rem;
          margin: 0 0 1rem 0;
        }

        .dev-status {
          display: flex;
          gap: 0.75rem;
          flex-wrap: wrap;
        }

        .status-badge {
          background: #e7f3ff;
          color: #0366d6;
          padding: 0.25rem 0.75rem;
          border-radius: 12px;
          font-size: 0.875rem;
          font-weight: 500;
          border: 1px solid #c1d9ff;
        }

        .component-container {
          flex: 1;
          display: flex;
          width: 100%;
          overflow: hidden;
        }

        daily-calendar {
          width: 100%;
          height: 100%;
        }

        @media (max-width: 1024px) {
          .dev-header {
            padding: 1rem;
          }

          .dev-header h1 {
            font-size: 1.5rem;
          }

          .dev-status {
            flex-direction: column;
            align-items: flex-start;
          }
        }
      </style>
    `;

    const existingStyles = document.getElementById('calendar-dev-page-styles');
    if (existingStyles) {
      existingStyles.remove();
    }
    document.head.insertAdjacentHTML('beforeend', styles);
  }

  private async loadCalendarComponent(): Promise<void> {
    try {
      await import('../calendar/daily-calendar.js');
    } catch (error) {
      this.showComponentError();
    }
  }

  private showComponentError(): void {
    const componentContainer = this.container.querySelector('.component-container');
    if (componentContainer) {
      componentContainer.innerHTML = `
        <div style="display: flex; align-items: center; justify-content: center; width: 100%; padding: 2rem; text-align: center;">
          <div>
            <h2 style="color: #dc2626; margin-bottom: 1rem;">Component Load Error</h2>
            <p style="color: #6b7280;">Failed to load the daily-calendar component.</p>
            <p style="color: #6b7280; font-size: 0.875rem;">Check the browser console for details.</p>
            <button onclick="window.location.reload()" style="margin-top: 1rem; padding: 0.5rem 1rem; background: #3b82f6; color: white; border: none; border-radius: 4px; cursor: pointer;">
              Reload Page
            </button>
          </div>
        </div>
      `;
    }
  }

  override destroy(): void {
    const styles = document.getElementById('calendar-dev-page-styles');
    if (styles) {
      styles.remove();
    }

    super.destroy();
  }
}

export default CalendarDevPage;