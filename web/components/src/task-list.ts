import { ComponentConfig } from './types.js';
import { TaskCard, Task } from './task-card.js';
import { TaskService, TasksResponse } from './task-service.js';

// Use global Sortable from CDN
declare const Sortable: any;

export class TaskList {
  private config: ComponentConfig;
  private container: HTMLElement;
  public taskCards: Map<string, TaskCard> = new Map();
  private sortableInstances: Map<string, any> = new Map();
  private tasksData: TasksResponse | null = null;
  private boundHandleClick?: (e: Event) => void;
  private boundHandleSubmit?: (e: Event) => void;
  private isSubmitting: boolean = false;
  
  // Separated responsibilities
  private taskService: TaskService;

  constructor(container: HTMLElement, config: ComponentConfig) {
    this.container = container;
    this.config = config;
    this.taskService = new TaskService(config);
    this.init();
  }

  private init(): void {
    this.container.className = 'task-list-container';
    this.loadTasks();
  }

  private async loadTasks(): Promise<void> {
    try {
      this.showLoading();
      this.tasksData = await this.taskService.getTasks();
      this.renderTasks();
    } catch (error) {
      this.showError('Failed to load tasks');
    }
  }

  private showLoading(): void {
    this.container.innerHTML = `
      <div class="loading-container">
        <div class="loading-spinner"></div>
        <p>Loading tasks...</p>
      </div>
    `;
  }

  private showError(message: string): void {
    // Create temporary error notification instead of replacing entire container
    const errorDiv = document.createElement('div');
    errorDiv.className = 'error-notification';
    errorDiv.style.cssText = `
      position: fixed;
      top: 20px;
      right: 20px;
      background: #f56565;
      color: white;
      padding: 12px 16px;
      border-radius: 4px;
      z-index: 1000;
      max-width: 300px;
    `;
    errorDiv.textContent = message;
    
    document.body.appendChild(errorDiv);
    
    // Auto-remove after 5 seconds
    setTimeout(() => {
      if (errorDiv.parentNode) {
        errorDiv.parentNode.removeChild(errorDiv);
      }
    }, 5000);
  }

  private renderTasks(): void {
    if (!this.tasksData) return;

    this.container.innerHTML = `
      <div class="task-list-header">
        <h2>Tasks for ${this.tasksData.date}</h2>
        <button class="add-task-btn" data-action="add-task">
          + Add Task
        </button>
      </div>
      <div class="task-columns">
        ${this.renderTaskColumns()}
      </div>
      ${this.renderAddTaskModal()}
    `;

    this.attachEventListeners();
    this.setupSortable();
  }

  private renderTaskColumns(): string {
    if (!this.tasksData) return '';

    const userEntries = Object.entries(this.tasksData.tasks_by_user);
    
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
          <span class="task-count">${column.tasks.length} tasks</span>
        </div>
        <div class="task-column-content" data-user-tasks="${userKey}">
          ${this.renderUserTasks(column.tasks)}
        </div>
      </div>
    `).join('');
  }

  private renderUserTasks(tasks: Task[]): string {
    return tasks.map(task => `<div data-task-container="${task.id}"></div>`).join('');
  }

  private renderAddTaskModal(): string {
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
                  ${this.renderAssigneeOptions()}
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

  private renderAssigneeOptions(): string {
    if (!this.tasksData) return '';
    
    return Object.values(this.tasksData.tasks_by_user)
      .map(column => `<option value="${column.user.id}">${column.user.name}</option>`)
      .join('');
  }

  private attachEventListeners(): void {
    this.boundHandleClick = this.handleClick.bind(this);
    this.boundHandleSubmit = this.handleSubmit.bind(this);
    
    this.container.addEventListener('click', this.boundHandleClick);
    this.container.addEventListener('submit', this.boundHandleSubmit);
    
    // Create TaskCard instances for each task
    this.createTaskCards();
  }

  private createTaskCards(): void {
    if (!this.tasksData) return;

    // Clear existing cards
    this.taskCards.forEach(card => card.destroy());
    this.taskCards.clear();

    Object.values(this.tasksData.tasks_by_user).forEach(column => {
      column.tasks.forEach(task => {
        const container = this.container.querySelector(`[data-task-container="${task.id}"]`) as HTMLElement;
        if (container) {
          const taskCard = new TaskCard(task, this.config, {
            onTaskUpdate: this.handleTaskUpdate.bind(this),
            onTaskDelete: this.handleTaskDelete.bind(this)
          });
          container.appendChild(taskCard.getElement());
          this.taskCards.set(task.id, taskCard);
        }
      });
    });
  }

  private setupSortable(): void {
    // Clear existing sortable instances
    this.sortableInstances.forEach(sortable => sortable.destroy());
    this.sortableInstances.clear();

    // Create sortable for each task column
    const taskColumns = this.container.querySelectorAll('[data-user-tasks]');
    taskColumns.forEach(column => {
      const userKey = column.getAttribute('data-user-tasks');
      if (userKey) {
        const sortable = new Sortable(column as HTMLElement, {
          group: 'tasks',
          animation: 150,
          ghostClass: 'task-ghost',
          chosenClass: 'task-chosen',
          dragClass: 'task-drag',
          onEnd: (evt: any) => this.handleTaskReorder(evt)
        });
        this.sortableInstances.set(userKey, sortable);
      }
    });
  }

  private handleClick(e: Event): void {
    const target = e.target as HTMLElement;
    const action = target.getAttribute('data-action');

    switch (action) {
      case 'add-task':
        this.showAddTaskModal();
        break;
      case 'close-modal':
        this.hideAddTaskModal();
        break;
    }
  }

  private handleSubmit(e: Event): void {
    e.preventDefault();
    const form = e.target as HTMLFormElement;
    
    if (form.getAttribute('data-form') === 'add-task') {
      this.handleAddTask(form);
    }
  }

  private showAddTaskModal(): void {
    const modal = this.container.querySelector('#add-task-modal') as HTMLElement;
    if (modal) {
      modal.style.display = 'flex';
      const titleInput = modal.querySelector('#task-title') as HTMLInputElement;
      titleInput?.focus();
    }
  }

  private hideAddTaskModal(): void {
    const modal = this.container.querySelector('#add-task-modal') as HTMLElement;
    if (modal) {
      modal.style.display = 'none';
      const form = modal.querySelector('.task-form') as HTMLFormElement;
      form?.reset();
    }
  }

  private async handleAddTask(form: HTMLFormElement): Promise<void> {
    if (this.isSubmitting) {
      return; // Prevent duplicate submissions
    }

    this.isSubmitting = true;
    
    const formData = new FormData(form);
    const taskData = {
      title: formData.get('title') as string,
      description: formData.get('description') as string,
      task_type: formData.get('task_type') as string,
      assigned_to: formData.get('assigned_to') as string || undefined,
      family_id: 'fam1' // Default family
    };

    try {
      await this.taskService.createTask(taskData);
      this.hideAddTaskModal();
      await this.loadTasks(); // Reload to get updated data
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to create task';
      this.showFormError(form, errorMessage);
    } finally {
      this.isSubmitting = false;
    }
  }

  private showFormError(form: HTMLFormElement, errors: any): void {
    // Clear existing errors
    form.querySelectorAll('.field-error').forEach(el => el.remove());

    if (typeof errors === 'string') {
      const errorDiv = document.createElement('div');
      errorDiv.className = 'field-error';
      errorDiv.textContent = errors;
      form.appendChild(errorDiv);
    } else if (Array.isArray(errors)) {
      errors.forEach(error => {
        const field = form.querySelector(`[name="${error.field}"]`);
        if (field) {
          const errorDiv = document.createElement('div');
          errorDiv.className = 'field-error';
          errorDiv.textContent = error.message;
          field.parentNode?.appendChild(errorDiv);
        }
      });
    }
  }

  private async handleTaskReorder(evt: any): Promise<void> {
    const taskElement = evt.item;
    const taskContainer = taskElement.querySelector('[data-task-id]');
    const taskId = taskContainer?.getAttribute('data-task-id');
    
    if (!taskId) return;

    // Check if task was moved between different columns (assignment change)
    if (evt.from !== evt.to) {
      const sourceColumn = evt.from.closest('[data-user-tasks]');
      const targetColumn = evt.to.closest('[data-user-tasks]');
      
      if (!targetColumn) {
        return;
      }

      const sourceUserKey = sourceColumn?.getAttribute('data-user-tasks');
      const targetUserKey = targetColumn.getAttribute('data-user-tasks');

      // Extract user ID from user key with validation
      let newAssignedTo: string | null = null;
      if (targetUserKey === 'unassigned') {
        newAssignedTo = null;
      } else if (targetUserKey?.startsWith('user_')) {
        const userId = targetUserKey.substring(5); // Remove 'user_' prefix
        if (userId && userId.trim() !== '') {
          newAssignedTo = userId;
        } else {
          this.showError(`Invalid user key format: ${targetUserKey}`);
          return;
        }
      } else {
        this.showError(`Unexpected user key format: ${targetUserKey}`);
        return;
      }

      // Store original state before making changes
      const originalTaskCounts = {
        source: sourceUserKey ? this.getTaskCount(sourceUserKey) : 0,
        target: this.getTaskCount(targetUserKey)
      };

      // Update UI optimistically
      this.updateTaskCounts(sourceUserKey, targetUserKey);
      this.updateTaskCardAssignment(taskId, newAssignedTo);

      try {
        await this.taskService.updateTask(taskId, { assigned_to: newAssignedTo });

      } catch (error) {
        // Comprehensive revert on error
        evt.from.insertBefore(evt.item, evt.from.children[evt.oldIndex] ?? null);
        
        // Revert task counts
        if (sourceUserKey) {
          this.setTaskCount(sourceUserKey, originalTaskCounts.source);
        }
        this.setTaskCount(targetUserKey, originalTaskCounts.target);
        
        // Revert task card assignment
        const originalAssignedTo = sourceUserKey === 'unassigned' ? null : 
                                  sourceUserKey?.substring(5) || null;
        this.updateTaskCardAssignment(taskId, originalAssignedTo);
        
        // Show error to user instead of throwing
        this.showError('Failed to update task assignment. Please try again.');
      }
    }
  }

  private getTaskCount(userKey: string): number {
    const countElement = this.container.querySelector(`[data-user-key="${userKey}"] .task-count`);
    if (!countElement) return 0;
    const match = countElement.textContent?.match(/(\d+)/);
    return match ? parseInt(match[1] || '0') : 0;
  }

  private setTaskCount(userKey: string, count: number): void {
    const countElement = this.container.querySelector(`[data-user-key="${userKey}"] .task-count`);
    if (countElement) {
      countElement.textContent = `${count} tasks`;
    }
  }

  private updateTaskCounts(sourceUserKey: string | null, targetUserKey: string): void {
    // Update task count for source user (decrease)
    if (sourceUserKey) {
      const currentCount = this.getTaskCount(sourceUserKey);
      this.setTaskCount(sourceUserKey, Math.max(0, currentCount - 1));
    }

    // Update task count for target user (increase)
    const currentCount = this.getTaskCount(targetUserKey);
    this.setTaskCount(targetUserKey, currentCount + 1);
  }

  private updateTaskCardAssignment(taskId: string, newAssignedTo: string | null): void {
    const taskCard = this.taskCards.get(taskId);
    if (taskCard) {
      taskCard.updateAssignment(newAssignedTo);
    }
  }

  private handleTaskUpdate(_task: Task): void {
    // Task was updated, we could update local state here if needed
    // For now, the TaskCard handles its own updates
  }

  private handleTaskDelete(taskId: string): void {
    const taskCard = this.taskCards.get(taskId);
    if (taskCard) {
      taskCard.destroy();
      this.taskCards.delete(taskId);
    }
  }

  public async refresh(): Promise<void> {
    await this.loadTasks();
  }

  public destroy(): void {
    this.taskCards.forEach(card => card.destroy());
    this.taskCards.clear();
    this.sortableInstances.forEach(sortable => sortable.destroy());
    this.sortableInstances.clear();
    
    // Remove event listeners
    if (this.boundHandleClick) {
      this.container.removeEventListener('click', this.boundHandleClick);
    }
    if (this.boundHandleSubmit) {
      this.container.removeEventListener('submit', this.boundHandleSubmit);
    }
  }
}