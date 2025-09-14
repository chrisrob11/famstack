import { ComponentConfig } from '../common/types.js';
import { TaskService, TasksResponse } from './task-service.js';
import { Task } from './task-types.js';

/**
 * DailyTasks - Component for showing daily tasks organized by family member
 * Used on the daily view page
 */
export class DailyTasks {
  private container: HTMLElement;
  private config: ComponentConfig;
  private taskService: TaskService;
  private tasks: TasksResponse | null = null;
  private boundHandleClick?: (e: Event) => void;

  constructor(container: HTMLElement, config: ComponentConfig) {
    this.container = container;
    this.config = config;
    this.taskService = new TaskService(config);
    this.init();
  }

  private init(): void {
    this.setupEventListeners();
    this.loadTasks();
  }

  private setupEventListeners(): void {
    this.boundHandleClick = this.handleClick.bind(this);
    this.container.addEventListener('click', this.boundHandleClick);
  }

  private async loadTasks(): Promise<void> {
    try {
      this.renderLoading();
      const tasksData = await this.taskService.getTasks();
      this.tasks = tasksData;
      this.renderTasks();
    } catch (error) {
      this.renderError('Failed to load tasks');
    }
  }

  private renderLoading(): void {
    this.container.innerHTML = `
      <div class="daily-tasks-loading">
        <div class="loading-spinner"></div>
        <p>Loading tasks...</p>
      </div>
    `;
  }

  private renderError(message: string): void {
    this.container.innerHTML = `
      <div class="daily-tasks-error">
        <p class="error-message">${message}</p>
        <button class="retry-btn" data-action="retry">Try Again</button>
      </div>
    `;
  }

  private renderTasks(): void {
    if (!this.tasks) {
      this.renderError('No tasks found');
      return;
    }

    this.container.innerHTML = `
      <div class="daily-tasks">
        <div class="daily-tasks-header">
          <h2>Daily Chores</h2>
          <button class="add-task-btn" data-action="add-task">
            <span class="add-icon">+</span>
          </button>
        </div>
        <div class="daily-tasks-grid">
          ${this.renderTasksGrid()}
        </div>
      </div>
    `;
  }

  private renderTasksGrid(): string {
    if (!this.tasks) return '';

    const userColumns = Object.values(this.tasks.tasks_by_user);
    const gridItems: string[] = [];

    userColumns.forEach(column => {
      const user = column.user;
      const tasks = column.tasks || [];

      // Filter to today's tasks (all pending tasks)
      const todayTasks = tasks.filter(task => task.status === 'pending');

      if (user.name !== 'Unassigned') {
        gridItems.push(`
          <div class="task-grid-item">
            <h3 class="user-name">${user.name}</h3>
            <div class="task-list">
              ${todayTasks.map(task => this.renderGridTaskItem(task)).join('')}
              ${todayTasks.length === 0 ? '<div class="no-tasks">No tasks</div>' : ''}
            </div>
          </div>
        `);
      }
    });

    return gridItems.join('');
  }

  private renderGridTaskItem(task: Task): string {
    return `
      <div class="grid-task-item" data-task-id="${task.id}">
        <label class="grid-task-label">
          <input 
            type="checkbox" 
            class="grid-task-checkbox" 
            data-action="toggle-task"
            data-task-id="${task.id}"
            ${task.status === 'completed' ? 'checked' : ''}
          >
          <div class="grid-task-content">
            <span class="grid-task-text">${task.title}</span>
            ${
              task.description && task.description !== task.title
                ? `<span class="grid-task-description">(${task.description})</span>`
                : ''
            }
          </div>
        </label>
      </div>
    `;
  }

  private renderTaskItem(task: Task): string {
    return `
      <div class="daily-task-item" data-task-id="${task.id}">
        <label class="task-checkbox-label">
          <input 
            type="checkbox" 
            class="task-checkbox" 
            data-action="toggle-task"
            data-task-id="${task.id}"
            ${task.status === 'completed' ? 'checked' : ''}
          >
          <span class="task-content">
            <span class="task-title">${task.title}</span>
            ${task.description ? `<span class="task-description">${task.description}</span>` : ''}
          </span>
        </label>
      </div>
    `;
  }

  private handleClick(e: Event): void {
    const target = e.target as HTMLElement;
    const action = target.getAttribute('data-action');

    switch (action) {
      case 'add-task':
        this.handleAddTask();
        break;
      case 'toggle-task':
        this.handleToggleTask(target);
        break;
      case 'retry':
        this.loadTasks();
        break;
    }
  }

  private handleAddTask(): void {
    // Emit event for parent component to handle
    this.container.dispatchEvent(
      new CustomEvent('add-task-requested', {
        bubbles: true,
        detail: { source: 'daily-tasks' },
      })
    );
  }

  private async handleToggleTask(checkboxElement: HTMLElement): Promise<void> {
    const taskId = checkboxElement.getAttribute('data-task-id');
    if (!taskId) return;

    const checkbox = checkboxElement as HTMLInputElement;
    const originalState = !checkbox.checked;

    try {
      const newStatus = checkbox.checked ? 'completed' : 'pending';
      await this.taskService.updateTask(taskId, { status: newStatus });

      // Update local state
      this.updateTaskStatus(taskId, newStatus);

      // Re-render to update UI
      this.renderTasks();
    } catch (error) {
      // Revert checkbox state on error
      checkbox.checked = originalState;
    }
  }

  private updateTaskStatus(taskId: string, status: string): void {
    if (!this.tasks) return;

    Object.values(this.tasks.tasks_by_user).forEach(column => {
      const task = column.tasks?.find(t => t.id === taskId);
      if (task) {
        task.status = status as 'pending' | 'completed';
        if (status === 'completed') {
          task.completed_at = new Date().toISOString();
        } else {
          task.completed_at = undefined;
        }
      }
    });
  }

  public async refresh(): Promise<void> {
    await this.loadTasks();
  }

  public destroy(): void {
    if (this.boundHandleClick) {
      this.container.removeEventListener('click', this.boundHandleClick);
    }
  }
}
