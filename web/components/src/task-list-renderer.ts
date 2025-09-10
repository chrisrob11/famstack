import { TaskColumn, TasksResponse } from './task-service.js';

/**
 * TaskListRenderer handles rendering the task list HTML
 * Separation of concerns: UI rendering separate from business logic
 */
export class TaskListRenderer {
  renderTaskList(data: TasksResponse): string {
    if (!data || Object.keys(data.tasks_by_user).length === 0) {
      return this.renderEmptyState();
    }

    const columns = Object.entries(data.tasks_by_user)
      .map(([key, column]) => this.renderTaskColumn(key, column))
      .join('');

    return `
      <div class="task-date">${data.date}</div>
      <div class="task-columns">
        ${columns}
      </div>
      ${this.renderAddTaskModal()}
    `;
  }

  private renderTaskColumn(userKey: string, column: TaskColumn): string {
    const taskList = column.tasks.map(task => this.renderTaskContainer(task)).join('');
    const taskCount = column.tasks.length;

    return `
      <div class="task-column" data-user-key="${userKey}" data-user-tasks="${column.user.id}">
        <div class="task-column-header">
          <h3 class="task-column-title">${column.user.name}</h3>
          <span class="task-count">${taskCount} tasks</span>
        </div>
        <div class="task-list" data-user-id="${column.user.id}">
          ${taskList}
          ${taskCount === 0 ? this.renderFirstTaskPrompt() : ''}
        </div>
        <button class="add-task-btn" data-action="add-task">
          + Add Task
        </button>
      </div>
    `;
  }

  private renderTaskContainer(task: any): string {
    return `
      <div class="task-item" data-task-id="${task.id}" data-task-container="${task.id}">
        <!-- Task card content will be rendered by TaskCard component -->
      </div>
    `;
  }

  private renderFirstTaskPrompt(): string {
    return `
      <div class="empty-task-list">
        <p>No tasks yet</p>
        <button class="add-task-btn" data-action="add-task">
          Add your first task
        </button>
      </div>
    `;
  }

  private renderEmptyState(): string {
    return `
      <div class="empty-state">
        <h2>No tasks found</h2>
        <p>Create your first task to get started.</p>
      </div>
      ${this.renderAddTaskModal()}
    `;
  }

  private renderAddTaskModal(): string {
    return `
      <div class="task-modal" id="add-task-modal" style="display: none;">
        <div class="task-modal-content">
          <div class="task-modal-header">
            <h3>Add New Task</h3>
            <button class="task-modal-close" data-action="close-modal">&times;</button>
          </div>
          <form class="task-form" data-form="add-task">
            <div class="form-group">
              <label for="task-title">Title *</label>
              <input type="text" id="task-title" name="title" required>
            </div>
            <div class="form-group">
              <label for="task-description">Description</label>
              <textarea id="task-description" name="description" rows="3"></textarea>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label for="task-type">Type</label>
                <select id="task-type" name="task_type">
                  <option value="todo">Todo</option>
                  <option value="chore">Chore</option>
                  <option value="appointment">Appointment</option>
                </select>
              </div>
              <div class="form-group">
                <label for="assigned-to">Assign to</label>
                <select id="assigned-to" name="assigned_to">
                  <option value="">Unassigned</option>
                  ${this.renderAssignmentOptions()}
                </select>
              </div>
            </div>
            <div class="form-actions">
              <button type="button" data-action="close-modal">Cancel</button>
              <button type="submit">
                Add Task
              </button>
            </div>
          </form>
        </div>
      </div>
    `;
  }

  private renderAssignmentOptions(): string {
    // This would ideally get user data from the TasksResponse
    // For now, return empty - will be populated dynamically
    return '';
  }

  updateAssignmentOptions(columns: TaskColumn[]): void {
    const select = document.querySelector('#assigned-to') as HTMLSelectElement;
    if (!select) return;

    const options = columns
      .filter(column => column.user.id !== 'unassigned')
      .map(column => `<option value="${column.user.id}">${column.user.name}</option>`)
      .join('');

    // Keep the unassigned option and add user options
    const unassignedOption = select.querySelector('option[value=""]');
    select.innerHTML = '';
    if (unassignedOption) {
      select.appendChild(unassignedOption);
    }
    select.insertAdjacentHTML('beforeend', options);
  }
}