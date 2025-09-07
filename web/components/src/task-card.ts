import { ComponentConfig } from './types.js';

export interface Task {
  id: string;
  title: string;
  description: string;
  status: 'pending' | 'completed';
  task_type: 'todo' | 'chore' | 'appointment';
  assigned_to?: string | null;
  created_at: string;
  completed_at?: string;
  priority: number;
}

export class TaskCard {
  private config: ComponentConfig;
  private task: Task;
  private element: HTMLElement;
  private onTaskUpdate: ((task: Task) => void) | undefined;
  private onTaskDelete: ((taskId: string) => void) | undefined;
  private boundHandleClick?: (e: Event) => void;
  private boundHandleDoubleClick?: (e: Event) => void;

  constructor(task: Task, config: ComponentConfig, options?: {
    onTaskUpdate?: (task: Task) => void;
    onTaskDelete?: (taskId: string) => void;
  }) {
    this.task = task;
    this.config = config;
    this.onTaskUpdate = options?.onTaskUpdate;
    this.onTaskDelete = options?.onTaskDelete;
    this.element = this.createElement();
    this.attachEventListeners();
  }

  private createElement(): HTMLElement {
    const cardElement = document.createElement('div');
    cardElement.className = `task-card ${this.task.status === 'completed' ? 'completed' : ''}`;
    cardElement.setAttribute('data-task-id', this.task.id);
    cardElement.innerHTML = this.getCardHTML();
    return cardElement;
  }

  private getCardHTML(): string {
    const isCompleted = this.task.status === 'completed';
    const completedAt = this.task.completed_at ? new Date(this.task.completed_at).toLocaleDateString() : '';
    
    return `
      <div class="task-header">
        <div class="task-title ${isCompleted ? 'completed' : ''}" data-editable>
          ${this.task.title}
        </div>
        <div class="task-actions">
          ${!isCompleted ? `
            <button class="task-action-btn" data-action="complete" style="color: green;" title="Complete">
              ✓
            </button>
          ` : `
            <button class="task-action-btn" data-action="reopen" style="color: blue;" title="Reopen">
              ↻
            </button>
          `}
          <button class="task-action-btn" data-action="edit" style="color: gray;" title="Edit">
            ✎
          </button>
          <button class="task-action-btn" data-action="delete" style="color: red;" title="Delete">
            ×
          </button>
        </div>
      </div>
      ${this.task.description ? `
        <div class="task-description">
          ${this.task.description}
        </div>
      ` : ''}
      <div class="task-meta">
        <span class="task-type ${this.task.task_type}">${this.task.task_type}</span>
        ${this.task.assigned_to ? `<span class="task-assignee">@${this.task.assigned_to}</span>` : ''}
        ${isCompleted && completedAt ? `<span class="completed-date">Completed ${completedAt}</span>` : ''}
      </div>
    `;
  }


  private attachEventListeners(): void {
    this.boundHandleClick = this.handleClick.bind(this);
    this.boundHandleDoubleClick = this.handleDoubleClick.bind(this);
    
    this.element.addEventListener('click', this.boundHandleClick);
    this.element.addEventListener('dblclick', this.boundHandleDoubleClick);
  }

  private handleClick(e: Event): void {
    const target = e.target as HTMLElement;
    const action = target.getAttribute('data-action');

    if (action === 'menu') {
      this.toggleMenu();
    } else if (action) {
      this.closeMenu();
      this.handleAction(action);
    } else {
      this.closeMenu();
    }
  }

  private handleDoubleClick(e: Event): void {
    const target = e.target as HTMLElement;
    if (target.hasAttribute('data-editable')) {
      this.enableInlineEdit(target);
    }
  }


  private toggleMenu(): void {
    const menu = this.element.querySelector('.task-menu') as HTMLElement;
    if (menu) {
      const isVisible = menu.style.display !== 'none';
      menu.style.display = isVisible ? 'none' : 'block';
      
      // Add/remove menu-open class to control z-index
      if (isVisible) {
        this.element.classList.remove('menu-open');
      } else {
        this.element.classList.add('menu-open');
      }
    }
  }

  private closeMenu(): void {
    const menu = this.element.querySelector('.task-menu') as HTMLElement;
    if (menu) {
      menu.style.display = 'none';
      this.element.classList.remove('menu-open');
    }
  }

  private async handleAction(action: string): Promise<void> {
    switch (action) {
      case 'complete':
        await this.updateTaskStatus('completed');
        break;
      case 'reopen':
        await this.updateTaskStatus('pending');
        break;
      case 'edit':
        this.enableInlineEdit(this.element.querySelector('.task-title') as HTMLElement);
        break;
      case 'delete':
        await this.deleteTask();
        break;
    }
  }

  private async updateTaskStatus(status: 'pending' | 'completed'): Promise<void> {
    try {
      const response = await fetch(`${this.config.apiBaseUrl}/tasks/${this.task.id}`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': this.config.csrfToken,
        },
        body: JSON.stringify({ status }),
      });

      if (response.ok) {
        const updatedTask = await response.json();
        this.updateTask(updatedTask);
      } else {
        throw new Error(`Failed to update task status: ${response.statusText}`);
      }
    } catch (error) {
      this.showError('Failed to update task status');
    }
  }

  private async deleteTask(): Promise<void> {
    if (!confirm('Are you sure you want to delete this task?')) {
      return;
    }
    
    try {
      const response = await fetch(`${this.config.apiBaseUrl}/tasks/${this.task.id}`, {
        method: 'DELETE',
        headers: {
          'X-CSRF-Token': this.config.csrfToken,
        },
      });

      if (response.ok) {
        this.onTaskDelete?.(this.task.id);
        this.element.remove();
      } else {
        throw new Error(`Failed to delete task: ${response.statusText}`);
      }
    } catch (error) {
      this.showError('Failed to delete task');
    }
  }

  private enableInlineEdit(titleElement: HTMLElement): void {
    const currentText = titleElement.textContent?.trim() || '';
    
    const input = document.createElement('input');
    input.type = 'text';
    input.value = currentText;
    input.className = 'task-title-input';

    const saveEdit = async () => {
      const newTitle = input.value.trim();
      if (newTitle === currentText || !newTitle) {
        titleElement.style.display = '';
        input.remove();
        return;
      }

      try {
        const response = await fetch(`${this.config.apiBaseUrl}/tasks/${this.task.id}`, {
          method: 'PATCH',
          headers: {
            'Content-Type': 'application/json',
            'X-CSRF-Token': this.config.csrfToken,
          },
          body: JSON.stringify({ title: newTitle }),
        });

        if (response.ok) {
          const updatedTask = await response.json();
          this.updateTask(updatedTask);
        } else {
          throw new Error('Failed to update task title');
        }
      } catch (error) {
        this.showError('Failed to update task title');
        titleElement.textContent = currentText;
      } finally {
        titleElement.style.display = '';
        input.remove();
      }
    };

    const cancelEdit = () => {
      titleElement.style.display = '';
      input.remove();
    };

    input.addEventListener('blur', saveEdit);
    input.addEventListener('keydown', (e) => {
      if (e.key === 'Enter') {
        e.preventDefault();
        input.blur();
      } else if (e.key === 'Escape') {
        e.preventDefault();
        cancelEdit();
      }
    });

    titleElement.style.display = 'none';
    titleElement.parentNode?.insertBefore(input, titleElement.nextSibling);
    input.focus();
    input.select();
  }

  private updateTask(updatedTask: Task): void {
    this.task = updatedTask;
    this.element.innerHTML = this.getCardHTML();
    this.element.className = `task-card ${this.task.status === 'completed' ? 'completed' : ''}`;
    this.onTaskUpdate?.(this.task);
  }

  private showError(message: string): void {
    const errorDiv = document.createElement('div');
    errorDiv.className = 'error-message';
    errorDiv.textContent = message;
    errorDiv.style.cssText = `
      position: fixed;
      top: 20px;
      right: 20px;
      background: #f56565;
      color: white;
      padding: 12px 16px;
      border-radius: 4px;
      z-index: 1000;
      box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    `;

    document.body.appendChild(errorDiv);
    setTimeout(() => errorDiv.remove(), 3000);
  }

  public getElement(): HTMLElement {
    return this.element;
  }

  public getTask(): Task {
    return this.task;
  }

  public updateAssignment(newAssignedTo: string | null): void {
    if (newAssignedTo) {
      this.task.assigned_to = newAssignedTo;
    } else {
      delete this.task.assigned_to;
    }
    this.element.innerHTML = this.getCardHTML();
    this.onTaskUpdate?.(this.task);
  }

  public destroy(): void {
    if (this.boundHandleClick) {
      this.element.removeEventListener('click', this.boundHandleClick);
    }
    if (this.boundHandleDoubleClick) {
      this.element.removeEventListener('dblclick', this.boundHandleDoubleClick);
    }
    this.element.remove();
  }
}