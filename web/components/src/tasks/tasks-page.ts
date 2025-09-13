import { BasePage } from '../pages/base-page.js';
import { TaskList } from './task-list.js';
import { ComponentConfig } from '../common/types.js';

/**
 * TasksPage - Complete owner of all task-related functionality
 * No more global component scanning - page creates and manages everything
 */
export class TasksPage extends BasePage {
  private taskList?: TaskList;

  constructor(container: HTMLElement, config: ComponentConfig) {
    super(container, config, 'tasks');
  }

  async init(): Promise<void> {
    try {
      // Create the basic page structure WITHOUT data-component attributes
      // This prevents any global scanning conflicts
      this.container.innerHTML = `
        <div class="tasks-page">
          <div class="tasks-container" id="main-task-container"></div>
        </div>
      `;

      // Page directly creates and owns the TaskList - no global scanning
      const taskContainer = this.container.querySelector('#main-task-container') as HTMLElement;
      this.taskList = new TaskList(taskContainer, this.config);

    } catch (error) {
      this.showError('Failed to load tasks page');
    }
  }

  async refresh(): Promise<void> {
    if (this.taskList) {
      await this.taskList.refresh();
    }
  }

  override destroy(): void {
    if (this.taskList) {
      this.taskList.destroy();
    }
    super.destroy();
  }
}