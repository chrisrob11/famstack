import { Task, ComponentConfig } from './types';
import Sortable from 'sortablejs';

export class TaskManager {
  private config: ComponentConfig;
  private container: HTMLElement;
  private sortable: Sortable | null = null;

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
    const taskList = this.container.querySelector('.task-list');
    if (!taskList) return;

    this.sortable = new Sortable(taskList as HTMLElement, {
      animation: 150,
      ghostClass: 'task-ghost',
      chosenClass: 'task-chosen',
      dragClass: 'task-drag',
      onEnd: (evt) => {
        this.handleTaskReorder(evt);
      },
    });
  }

  private setupInlineEditing(): void {
    this.container.addEventListener('dblclick', (e) => {
      const target = e.target as HTMLElement;
      const taskTitle = target.closest('.task-title');
      
      if (taskTitle) {
        this.enableInlineEdit(taskTitle as HTMLElement);
      }
    });
  }

  private setupTaskActions(): void {
    this.container.addEventListener('click', (e) => {
      const target = e.target as HTMLElement;
      
      if (target.matches('.task-complete-btn')) {
        this.handleTaskComplete(target);
      } else if (target.matches('.task-delete-btn')) {
        this.handleTaskDelete(target);
      }
    });
  }

  private async handleTaskReorder(evt: Sortable.SortableEvent): Promise<void> {
    if (evt.oldIndex === undefined || evt.newIndex === undefined) return;

    const taskElement = evt.item;
    const taskId = taskElement.getAttribute('data-task-id');
    
    if (!taskId) return;

    try {
      const response = await fetch(`${this.config.apiBaseUrl}/tasks/${taskId}/reorder`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': this.config.csrfToken,
        },
        body: JSON.stringify({
          oldIndex: evt.oldIndex,
          newIndex: evt.newIndex,
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to reorder task');
      }
    } catch (error) {
      console.error('Error reordering task:', error);
      // Revert the change on error
      if (this.sortable) {
        if (evt.oldIndex < evt.newIndex) {
          evt.to.insertBefore(evt.item, evt.to.children[evt.oldIndex] ?? null);
        } else {
          evt.to.insertBefore(evt.item, evt.to.children[evt.oldIndex + 1] ?? null);
        }
      }
    }
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
    input.addEventListener('keydown', (e) => {
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
      console.error('Error updating task:', error);
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
      console.error('Error completing task:', error);
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
      console.error('Error deleting task:', error);
      this.showError('Failed to delete task');
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
    if (this.sortable) {
      this.sortable.destroy();
      this.sortable = null;
    }
  }
}