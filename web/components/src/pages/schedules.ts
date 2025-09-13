import { ComponentConfig } from '../common/types.js';
import { ScheduleList } from '../schedules/schedule-list.js';

export class SchedulesPage {
  private config: ComponentConfig;
  private container: HTMLElement;
  private scheduleList: ScheduleList | null = null;

  constructor(config: ComponentConfig) {
    this.config = config;
    this.container = document.createElement('div');
    this.container.className = 'schedules-page';
  }

  public render(targetContainer: HTMLElement): void {
    targetContainer.appendChild(this.container);
    
    // Initialize the schedule list
    this.scheduleList = new ScheduleList(this.container, this.config);
  }

  public destroy(): void {
    if (this.scheduleList) {
      this.scheduleList.destroy();
      this.scheduleList = null;
    }
    this.container.remove();
  }

  public getElement(): HTMLElement {
    return this.container;
  }
}