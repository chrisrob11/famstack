import { BasePage } from './base-page.js';
import { ComponentConfig } from '../common/types.js';

export class CalendarDevPage extends BasePage {
  constructor(container: HTMLElement, config: ComponentConfig) {
    super(container, config, 'calendar-dev');
  }

  async init(): Promise<void> {
    this.render();
    this.loadCalendarComponent();
    this.setupEventListeners();
  }

  private render(): void {
    this.container.innerHTML = `
      <div class="calendar-dev-page">
        <div class="dev-header">
          <h1>📅 Daily Calendar Component - Development</h1>
          <p>Milestone 2: Core Layout & Time Grid - Full 24-hour coverage with configurable format</p>
          <div class="dev-status">
            <span class="status-badge">✅ Time Grid Complete</span>
            <span class="status-badge">✅ Scroll Synchronization</span>
            <span class="status-badge">🔄 12h/24h Time Format</span>
          </div>
          <div class="dev-controls">
            <label>
              <input type="checkbox" id="time-format-toggle"> Use 24-hour format
            </label>
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

        .dev-controls {
          margin-top: 1rem;
          padding-top: 1rem;
          border-top: 1px solid #e9ecef;
        }

        .dev-controls label {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          font-size: 0.875rem;
          color: #495057;
          cursor: pointer;
        }

        .dev-controls input[type="checkbox"] {
          width: 16px;
          height: 16px;
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

  private setupEventListeners(): void {
    const toggle = this.container.querySelector('#time-format-toggle') as HTMLInputElement;
    const calendar = this.container.querySelector('daily-calendar') as any;

    if (toggle && calendar) {
      toggle.addEventListener('change', () => {
        calendar.use24Hour = toggle.checked;
      });
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
