import { ComponentConfig } from '../common/types.js';
import { Task } from './task-types.js';

/**
 * PersonTasks - Component for displaying tasks for a single family member
 * Used within the FamilyTasks grid component
 */
export class PersonTasks {
  private container: HTMLElement;
  private config: ComponentConfig;
  private member: { id: string; name: string; member_type: string };
  private tasks: Task[] = [];
  private boundHandleClick?: (e: Event) => void;

  constructor(
    container: HTMLElement,
    config: ComponentConfig,
    member: { id: string; name: string; member_type: string }
  ) {
    this.container = container;
    this.config = config;
    this.member = member;
    this.init();
  }

  private init(): void {
    this.setupEventListeners();
    this.render();
  }

  private setupEventListeners(): void {
    this.boundHandleClick = this.handleClick.bind(this);
    this.container.addEventListener('click', this.boundHandleClick);
  }

  public setTasks(tasks: Task[]): void {
    // Filter to today's tasks (all pending tasks) and sort by created_at
    this.tasks = tasks
      .filter(task => task.status === 'pending')
      .sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime());
    this.render();
  }

  private render(): void {
    this.container.innerHTML = `
      <div class="person-tasks">
        <div class="person-header">
          <h3 class="person-name">${this.member.name}</h3>
          <button class="person-add-btn" data-action="add-task" data-member-id="${this.member.id}">
            <span class="person-add-icon">+</span>
          </button>
        </div>
        <div class="person-task-list">
          ${this.tasks.map(task => this.renderTask(task)).join('')}
        </div>
      </div>
    `;
  }

  private renderTask(task: Task): string {
    return `
      <div class="person-task-item" data-task-id="${task.id}">
        <div class="person-task-content">
          <input 
            type="checkbox" 
            class="person-task-checkbox" 
            data-action="toggle-task"
            data-task-id="${task.id}"
            ${task.status === 'completed' ? 'checked' : ''}
          >
          <span class="person-task-text" data-action="edit-task" data-task-id="${task.id}">${task.title}</span>
        </div>
      </div>
    `;
  }

  private handleClick(e: Event): void {
    const target = e.target as HTMLElement;
    const actionElement = target.closest('[data-action]') as HTMLElement;

    if (!actionElement) return;

    const action = actionElement.getAttribute('data-action');

    switch (action) {
      case 'toggle-task':
        this.handleToggleTask(actionElement);
        break;
      case 'add-task':
        this.handleAddTask(actionElement);
        break;
      case 'edit-task':
        this.handleEditTask(actionElement);
        break;
    }
  }

  private handleAddTask(element: HTMLElement): void {
    const memberId = element.getAttribute('data-member-id');
    if (!memberId) return;

    // Emit event to parent component for handling
    this.container.dispatchEvent(
      new CustomEvent('person-add-task', {
        bubbles: true,
        detail: {
          memberId,
          memberName: this.member.name,
        },
      })
    );
  }

  private handleEditTask(element: HTMLElement): void {
    const taskId = element.getAttribute('data-task-id');
    if (!taskId) return;

    const currentText = element.textContent || '';

    // Create input element
    const input = document.createElement('input');
    input.type = 'text';
    input.value = currentText;
    input.className = 'person-task-input';

    // Replace the span with the input
    element.parentNode?.replaceChild(input, element);
    input.focus();
    input.select();

    // Handle save on Enter or blur
    const saveEdit = async () => {
      const newText = input.value.trim();
      if (newText && newText !== currentText) {
        // Emit event to parent component for handling
        this.container.dispatchEvent(
          new CustomEvent('task-update', {
            bubbles: true,
            detail: {
              taskId,
              title: newText,
            },
          })
        );
      }

      // Create new span and replace input
      const newSpan = document.createElement('span');
      newSpan.className = 'person-task-text';
      newSpan.setAttribute('data-action', 'edit-task');
      newSpan.setAttribute('data-task-id', taskId);
      newSpan.textContent = newText || currentText;

      input.parentNode?.replaceChild(newSpan, input);
    };

    input.addEventListener('blur', saveEdit);
    input.addEventListener('keydown', e => {
      if (e.key === 'Enter') {
        saveEdit();
      } else if (e.key === 'Escape') {
        // Cancel editing
        const newSpan = document.createElement('span');
        newSpan.className = 'person-task-text';
        newSpan.setAttribute('data-action', 'edit-task');
        newSpan.setAttribute('data-task-id', taskId);
        newSpan.textContent = currentText;

        input.parentNode?.replaceChild(newSpan, input);
      }
    });
  }

  private async handleToggleTask(checkboxElement: HTMLElement): Promise<void> {
    const taskId = checkboxElement.getAttribute('data-task-id');
    if (!taskId) return;

    const checkbox = checkboxElement as HTMLInputElement;
    const originalState = !checkbox.checked;

    // Emit event to parent component for handling
    this.container.dispatchEvent(
      new CustomEvent('task-toggle', {
        bubbles: true,
        detail: {
          taskId,
          newStatus: checkbox.checked ? 'completed' : 'pending',
          originalState,
          checkbox,
        },
      })
    );
  }

  public updateTaskStatus(taskId: string, status: string): void {
    const task = this.tasks.find(t => t.id === taskId);
    if (task) {
      task.status = status as 'pending' | 'completed';
      if (status === 'completed') {
        task.completed_at = new Date().toISOString();
      } else {
        task.completed_at = null;
      }
    }
  }

  public destroy(): void {
    if (this.boundHandleClick) {
      this.container.removeEventListener('click', this.boundHandleClick);
    }
  }
}
