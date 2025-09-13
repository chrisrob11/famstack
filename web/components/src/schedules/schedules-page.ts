import { BasePage } from '../pages/base-page.js';
import { ScheduleList } from './schedule-list.js';
import { ComponentConfig } from '../common/types.js';

/**
 * SchedulesPage - Complete owner of all schedule-related functionality
 * No more global component scanning - page creates and manages everything
 */
export class SchedulesPage extends BasePage {
  private scheduleList?: ScheduleList;

  constructor(container: HTMLElement, config: ComponentConfig) {
    super(container, config, 'schedules');
  }

  async init(): Promise<void> {
    // Create the basic page structure WITHOUT data-component attributes
    // This prevents any global scanning conflicts
    this.container.innerHTML = `
      <div class="schedules-page">
        <div class="schedules-container" id="main-schedule-container"></div>
      </div>
    `;

    // Page directly creates and owns the ScheduleList - no global scanning
    const scheduleContainer = this.container.querySelector('#main-schedule-container') as HTMLElement;
    if (scheduleContainer) {
      this.scheduleList = new ScheduleList(scheduleContainer, this.config);
    }
  }

  async refresh(): Promise<void> {
    if (this.scheduleList) {
      // Reload schedules data
      await this.scheduleList.refresh();
    }
  }

  override destroy(): void {
    if (this.scheduleList) {
      this.scheduleList.destroy();
      delete this.scheduleList;
    }
    super.destroy();
  }
}