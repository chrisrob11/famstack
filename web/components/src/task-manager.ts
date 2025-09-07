import { ComponentConfig } from './types.js';

// Use global Sortable from CDN
declare const Sortable: any;

export class TaskManager {
  private config: ComponentConfig;
  private container: HTMLElement;
  private sortables: any[] = [];
  private boundHandleDblClick?: (e: Event) => void;
  private boundHandleClick?: (e: Event) => void;

  constructor(container: HTMLElement, config: ComponentConfig) {
    this.container = container;
    this.config = config;
    this.init();
  }

  private init(): void {
    this.setupDragAndDrop();
    this.setupInlineEditing();
    this.setupTaskActions();
  }

  private setupDragAndDrop(): void {
    // Set up drag and drop for all user columns
    const userColumns = this.container.querySelectorAll('[data-user-id]');
    
    userColumns.forEach(column => {
      const taskList = column.querySelector('.task-list');
      if (!taskList) return;

      const sortable = new Sortable(taskList as HTMLElement, {
        group: 'tasks', // Allow dragging between different columns
        animation: 150,
        ghostClass: 'task-ghost',
        chosenClass: 'task-chosen',
        dragClass: 'task-drag',
        onEnd: (evt: any) => {
          this.handleTaskReorder(evt);
        },
      });
      
      this.sortables.push(sortable);
    });
  }

  private setupInlineEditing(): void {
    this.boundHandleDblClick = (e: Event) => {
      const target = e.target as HTMLElement;
      const taskTitle = target.closest('.task-title');

      if (taskTitle) {
        this.enableInlineEdit(taskTitle as HTMLElement);
      }
    };
    
    this.container.addEventListener('dblclick', this.boundHandleDblClick);
  }

  private setupTaskActions(): void {
    this.boundHandleClick = (e: Event) => {
      const target = e.target as HTMLElement;

      if (target.matches('.task-complete-btn')) {
        this.handleTaskComplete(target);
      } else if (target.matches('.task-delete-btn')) {
        this.handleTaskDelete(target);
      }
    };
    
    this.container.addEventListener('click', this.boundHandleClick);
  }

  private async handleTaskReorder(evt: any): Promise<void> {
    const taskElement = evt.item;
    const taskId = taskElement.getAttribute('data-task-id');

    if (!taskId) {
      return;
    }

    // Check if task was moved between different columns (assignment change)
    if (evt.from !== evt.to) {
      const sourceColumn = evt.from.closest('[data-user-id]');
      const targetColumn = evt.to.closest('[data-user-id]');
      
      if (!targetColumn) {
        return;
      }

      const newAssignedTo = targetColumn.getAttribute('data-user-id');
      const oldAssignedTo = sourceColumn?.getAttribute('data-user-id');

      const assignmentValue = newAssignedTo === 'unassigned' ? null : newAssignedTo;

      try {
        // Update assignment in database
        const url = `${this.config.apiBaseUrl}/tasks/${taskId}`;
        const payload = { assigned_to: assignmentValue };
        
        const response = await fetch(url, {
          method: 'PATCH',
          headers: {
            'Content-Type': 'application/json',
            'X-CSRF-Token': this.config.csrfToken,
          },
          body: JSON.stringify(payload),
        });

        if (!response.ok) {
          const errorText = await response.text();
          throw new Error(`Failed to update task assignment: ${errorText}`);
        }

        await response.json();

        // Update task count UI
        this.updateTaskCounts(oldAssignedTo, newAssignedTo);

      } catch (error) {
        // Revert the drag operation on error
        evt.from.insertBefore(evt.item, evt.from.children[evt.oldIndex] ?? null);
        this.showError('Failed to update task assignment');
      }
    }
    // If moved within same column, no API call needed (just reordering)
  }

  private enableInlineEdit(element: HTMLElement): void {
    const currentText = element.textContent ?? '';
    const taskElement = element.closest('.task-item');
    const taskId = taskElement?.getAttribute('data-task-id');

    if (!taskId) return;

    const input = document.createElement('input');
    input.type = 'text';
    input.value = currentText;
    input.className = 'task-title-input';

    input.addEventListener('blur', () => this.saveInlineEdit(element, input, taskId));
    input.addEventListener('keydown', e => {
      if (e.key === 'Enter') {
        e.preventDefault();
        input.blur();
      } else if (e.key === 'Escape') {
        element.textContent = currentText;
        element.style.display = '';
        input.remove();
      }
    });

    element.style.display = 'none';
    element.parentNode?.insertBefore(input, element.nextSibling);
    input.focus();
    input.select();
  }

  private async saveInlineEdit(
    titleElement: HTMLElement,
    input: HTMLInputElement,
    taskId: string
  ): Promise<void> {
    const newTitle = input.value.trim();

    if (newTitle === titleElement.textContent) {
      titleElement.style.display = '';
      input.remove();
      return;
    }

    try {
      const response = await fetch(`${this.config.apiBaseUrl}/tasks/${taskId}`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': this.config.csrfToken,
        },
        body: JSON.stringify({ title: newTitle }),
      });

      if (response.ok) {
        titleElement.textContent = newTitle;
      } else {
        throw new Error('Failed to update task title');
      }
    } catch (error) {
      // Show error feedback to user
      this.showError('Failed to update task title');
    } finally {
      titleElement.style.display = '';
      input.remove();
    }
  }

  private async handleTaskComplete(button: HTMLElement): Promise<void> {
    const taskElement = button.closest('.task-item');
    const taskId = taskElement?.getAttribute('data-task-id');

    if (!taskId) return;

    try {
      const response = await fetch(`${this.config.apiBaseUrl}/tasks/${taskId}/complete`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': this.config.csrfToken,
        },
      });

      if (response.ok) {
        taskElement?.classList.add('task-completed');
        button.textContent = 'Completed';
        button.setAttribute('disabled', 'true');
      } else {
        throw new Error('Failed to complete task');
      }
    } catch (error) {
      this.showError('Failed to complete task');
    }
  }

  private async handleTaskDelete(button: HTMLElement): Promise<void> {
    const taskElement = button.closest('.task-item');
    const taskId = taskElement?.getAttribute('data-task-id');

    if (!taskId || !confirm('Are you sure you want to delete this task?')) {
      return;
    }

    try {
      const response = await fetch(`${this.config.apiBaseUrl}/tasks/${taskId}`, {
        method: 'DELETE',
        headers: {
          'X-CSRF-Token': this.config.csrfToken,
        },
      });

      if (response.ok) {
        taskElement?.remove();
      } else {
        throw new Error('Failed to delete task');
      }
    } catch (error) {
      this.showError('Failed to delete task');
    }
  }

  private updateTaskCounts(oldUserId: string | null, newUserId: string): void {
    // Update task count for source user (decrease)
    if (oldUserId) {
      const oldCountElement = document.querySelector(`[data-user-id="${oldUserId}"] .task-count`);
      if (oldCountElement) {
        const currentCount = parseInt(oldCountElement.textContent || '0');
        oldCountElement.textContent = Math.max(0, currentCount - 1).toString();
      }
    }

    // Update task count for target user (increase)
    const newCountElement = document.querySelector(`[data-user-id="${newUserId}"] .task-count`);
    if (newCountElement) {
      const currentCount = parseInt(newCountElement.textContent || '0');
      newCountElement.textContent = (currentCount + 1).toString();
    }
  }

  private showError(message: string): void {
    // Simple error display - could be enhanced with a proper notification system
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
    `;

    document.body.appendChild(errorDiv);

    setTimeout(() => {
      errorDiv.remove();
    }, 3000);
  }

  public destroy(): void {
    this.sortables.forEach(sortable => {
      if (sortable) {
        sortable.destroy();
      }
    });
    this.sortables = [];
    
    // Remove event listeners
    if (this.boundHandleDblClick) {
      this.container.removeEventListener('dblclick', this.boundHandleDblClick);
    }
    if (this.boundHandleClick) {
      this.container.removeEventListener('click', this.boundHandleClick);
    }
  }
}
