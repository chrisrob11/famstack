import { TasksResponse } from './task-service.js';
import { ComponentUtils } from '../common/component-utils.js';

/**
 * TaskListRenderer - Single responsibility: HTML generation and DOM manipulation
 * Separated from business logic and event handling
 */
export class TaskListRenderer {
  private container: HTMLElement;

  constructor(container: HTMLElement) {
    this.container = container;
  }

  renderLoading(message: string = 'Loading tasks...'): void {
    ComponentUtils.showLoading(this.container, message);
  }

  renderError(message: string): void {
    ComponentUtils.showError(message);
  }

  renderTaskList(data: TasksResponse): void {
    const html = `
      <div class="task-list-header">
        <h2>Tasks for ${data.date}</h2>
        <button class="add-task-btn" data-action="add-task">
          + Add Task
        </button>
      </div>
      <div class="task-columns">
        ${this.renderTaskColumns(data)}
      </div>
      ${this.renderAddTaskModal(data)}
    `;
    this.container.innerHTML = html;
  }

  private renderTaskColumns(data: TasksResponse): string {
    const userEntries = Object.entries(data.tasks_by_user);
    
    if (userEntries.length === 0) {
      return `
        <div class="empty-state">
          <p>No tasks for today</p>
          <button class="add-task-btn" data-action="add-task">
            Add your first task
          </button>
        </div>
      `;
    }

    return userEntries.map(([userKey, column]) => `
      <div class="task-column" data-user-key="${userKey}">
        <div class="task-column-header">
          <h3>${column.user.name}</h3>
          <span class="task-count">${(column.tasks || []).length} tasks</span>
        </div>
        <div class="task-column-content" data-user-tasks="${userKey}">
          ${this.renderUserTasks(column.tasks || [])}
        </div>
      </div>
    `).join('');
  }

  private renderUserTasks(tasks: any[]): string {
    const html = tasks.map(task => {
      return `<div data-task-container="${task.id}"></div>`;
    }).join('');
    return html;
  }

  private renderAddTaskModal(tasksData: TasksResponse): string {
    return `
      <div class="task-modal" id="add-task-modal" style="display: none;">
        <div class="task-modal-content">
          <div class="task-modal-header">
            <h3>Add New Task</h3>
            <button class="modal-close" data-action="close-modal">&times;</button>
          </div>
          <form class="task-form" data-form="add-task">
            <div class="form-group">
              <label for="task-title">Title *</label>
              <input type="text" id="task-title" name="title" required maxlength="255">
            </div>
            <div class="form-group">
              <label for="task-description">Description</label>
              <textarea id="task-description" name="description" rows="3" maxlength="1000"></textarea>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label for="task-type">Type *</label>
                <select id="task-type" name="task_type" required>
                  <option value="todo">Todo</option>
                  <option value="chore">Chore</option>
                  <option value="appointment">Appointment</option>
                </select>
              </div>
              <div class="form-group">
                <label for="task-assignee">Assign To</label>
                <select id="task-assignee" name="assigned_to">
                  <option value="">Unassigned</option>
                  ${this.renderAssigneeOptions(tasksData)}
                </select>
              </div>
            </div>
            <div class="form-actions">
              <button type="button" class="btn btn-secondary" data-action="close-modal">
                Cancel
              </button>
              <button type="submit" class="btn btn-primary">
                Add Task
              </button>
            </div>
          </form>
        </div>
      </div>
    `;
  }

  private renderAssigneeOptions(tasksData: TasksResponse): string {
    return Object.values(tasksData.tasks_by_user)
      .map(column => `<option value="${column.user.id}">${column.user.name}</option>`)
      .join('');
  }

  updateTaskCount(userKey: string, count: number): void {
    const countElement = this.container.querySelector(`[data-user-key="${userKey}"] .task-count`);
    if (countElement) {
      countElement.textContent = `${count} tasks`;
    }
  }

  getTaskCount(userKey: string): number {
    const countElement = this.container.querySelector(`[data-user-key="${userKey}"] .task-count`);
    if (!countElement) return 0;
    
    const match = countElement.textContent?.match(/(\d+)/);
    return match ? parseInt(match[1] || '0') : 0;
  }

  showAddTaskModal(): void {
    const modal = this.container.querySelector('#add-task-modal') as HTMLElement;
    if (modal) {
      modal.style.display = 'flex';
      const titleInput = modal.querySelector('#task-title') as HTMLInputElement;
      titleInput?.focus();
    }
  }

  hideAddTaskModal(): void {
    const modal = this.container.querySelector('#add-task-modal') as HTMLElement;
    if (modal) {
      modal.style.display = 'none';
      const form = modal.querySelector('.task-form') as HTMLFormElement;
      form?.reset();
    }
  }
}