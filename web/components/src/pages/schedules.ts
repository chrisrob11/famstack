/**
 * Schedules page component for SPA
 */

import { BasePage } from './base-page.js';
import { ComponentConfig } from '../common/types.js';
import { ScheduleList } from '../schedules/schedule-list.js';

export class SchedulesPage extends BasePage {
  private scheduleList: ScheduleList | null = null;

  constructor(container: HTMLElement, config: ComponentConfig) {
    super(container, config, 'schedules');
  }

  async init(): Promise<void> {
    this.render();
    this.setupEventListeners();
  }

  private render(): void {
    this.container.innerHTML = `
      <div class="schedules-page">
        <div class="schedules-header">
          <h1>Schedules</h1>
          <p>Manage recurring schedules and routines</p>
        </div>
        <div id="schedules-container"></div>
      </div>
    `;

    this.addStyles();
    this.initializeComponents();
  }

  private addStyles(): void {
    const styles = `
      <style id="schedules-page-styles">
        .schedules-page {
          padding: 2rem;
          max-width: 1200px;
          margin: 0 auto;
        }

        .schedules-header {
          margin-bottom: 2rem;
        }

        .schedules-header h1 {
          font-size: 2rem;
          font-weight: 700;
          color: #374151;
          margin: 0 0 0.5rem 0;
        }

        .schedules-header p {
          color: #6b7280;
          font-size: 1rem;
          margin: 0;
        }

        #schedules-container {
          min-height: 400px;
        }
      </style>
    `;

    // Remove existing styles
    const existingStyles = document.getElementById('schedules-page-styles');
    if (existingStyles) {
      existingStyles.remove();
    }

    // Add styles to head
    document.head.insertAdjacentHTML('beforeend', styles);
  }

  private setupEventListeners(): void {
    // Add any page-level event listeners here
  }

  private initializeComponents(): void {
    const schedulesContainer = this.container.querySelector('#schedules-container') as HTMLElement;
    if (schedulesContainer) {
      this.scheduleList = new ScheduleList(schedulesContainer, this.config);
    }
  }

  async refresh(): Promise<void> {
    if (this.scheduleList) {
      await this.scheduleList.refresh();
    }
  }

  override destroy(): void {
    if (this.scheduleList) {
      this.scheduleList.destroy();
      this.scheduleList = null;
    }

    // Remove styles
    const styles = document.getElementById('schedules-page-styles');
    if (styles) {
      styles.remove();
    }

    super.destroy();
  }
}

export default SchedulesPage;
