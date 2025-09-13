import { ComponentConfig } from '../common/types.js';
import { Task } from './task-types.js';

export interface TaskFormData {
  title: string;
  description: string;
  task_type: 'todo' | 'chore' | 'appointment';
  assigned_to?: string | null;
  priority: number;
}

export interface TaskModalOptions {
  onSave: (data: TaskFormData, taskId?: string) => Promise<void>;
  onCancel?: () => void;
  familyMembers?: Array<{ id: string; name: string }>;
}

export class TaskModal {
  private element!: HTMLElement;
  private backdrop!: HTMLElement;
  private currentTask: Task | undefined;
  private isSubmitting: boolean = false;
  private onSave: (data: TaskFormData, taskId?: string) => Promise<void>;
  private onCancel: (() => void) | undefined;
  private familyMembers: Array<{ id: string; name: string }>;
  private boundHandleClick?: (e: Event) => void;
  private boundHandleSubmit?: (e: Event) => void;

  constructor(container: HTMLElement, config: ComponentConfig, options: TaskModalOptions) {
    this.onSave = options.onSave;
    this.onCancel = options.onCancel;
    this.familyMembers = options.familyMembers || [];

    this.createElement(container);
    this.attachEventListeners();
  }

  private createElement(container: HTMLElement): void {
    // Create modal
    this.element = document.createElement('div');
    this.element.className = 'modal';
    this.element.id = 'task-modal';
    this.element.style.display = 'none';
    this.element.innerHTML = this.getModalHTML();

    // Create backdrop
    this.backdrop = document.createElement('div');
    this.backdrop.className = 'modal-backdrop';
    this.backdrop.style.display = 'none';

    container.appendChild(this.element);
    container.appendChild(this.backdrop);
  }

  private getModalHTML(): string {
    return `
      <div class="modal-content">
        <div class="modal-header">
          <h2 id="modal-title">Add Task</h2>
          <button type="button" class="modal-close" data-action="close">×</button>
        </div>
        
        <form id="task-form">
          <div class="modal-body">
            <div class="form-group">
              <label for="task-title" class="form-label">Task Title *</label>
              <input 
                type="text" 
                id="task-title" 
                name="title" 
                class="form-input" 
                required 
                placeholder="Enter task title"
                maxlength="100"
              >
              <div class="form-error" id="title-error"></div>
            </div>

            <div class="form-group">
              <label for="task-description" class="form-label">Description</label>
              <textarea 
                id="task-description" 
                name="description" 
                class="form-input" 
                placeholder="Enter task description (optional)"
                rows="3"
                maxlength="500"
              ></textarea>
              <div class="form-error" id="description-error"></div>
            </div>

            <div class="form-group">
              <label for="task-type" class="form-label">Task Type *</label>
              <select id="task-type" name="task_type" class="form-input" required>
                <option value="">Select task type</option>
                <option value="chore">Chore</option>
                <option value="todo">To-Do</option>
                <option value="appointment">Appointment</option>
              </select>
              <div class="form-error" id="task-type-error"></div>
            </div>

            <div class="form-group">
              <label for="assigned-to" class="form-label">Assign To</label>
              <select id="assigned-to" name="assigned_to" class="form-input">
                <option value="">Unassigned</option>
                ${this.familyMembers
                  .map(member => `<option value="${member.id}">${member.name}</option>`)
                  .join('')}
              </select>
              <div class="form-error" id="assigned-to-error"></div>
            </div>

            <div class="form-group">
              <label for="priority" class="form-label">Priority</label>
              <select id="priority" name="priority" class="form-input">
                <option value="0">Low</option>
                <option value="1" selected>Normal</option>
                <option value="2">High</option>
                <option value="3">Urgent</option>
              </select>
              <div class="form-error" id="priority-error"></div>
            </div>
          </div>
          
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-action="cancel">Cancel</button>
            <button type="submit" class="btn btn-primary" id="save-button">
              <span class="button-text">Save Task</span>
              <span class="button-loading" style="display: none;">Saving...</span>
            </button>
          </div>
        </form>
      </div>
    `;
  }

  private attachEventListeners(): void {
    this.boundHandleClick = this.handleClick.bind(this);
    this.boundHandleSubmit = this.handleSubmit.bind(this);

    this.element.addEventListener('click', this.boundHandleClick);
    this.element.addEventListener('submit', this.boundHandleSubmit);
    this.backdrop.addEventListener('click', () => this.hide());

    // Escape key to close modal
    document.addEventListener('keydown', e => {
      if (e.key === 'Escape' && this.element.style.display !== 'none') {
        this.hide();
      }
    });
  }

  private handleClick(e: Event): void {
    const target = e.target as HTMLElement;
    const action = target.getAttribute('data-action');

    switch (action) {
      case 'close':
      case 'cancel':
        this.hide();
        break;
    }
  }

  private async handleSubmit(e: Event): Promise<void> {
    e.preventDefault();

    if (this.isSubmitting) return;

    const form = e.target as HTMLFormElement;
    const formData = new FormData(form);

    // Clear previous errors
    this.clearErrors();

    const data: TaskFormData = {
      title: (formData.get('title') as string).trim(),
      description: (formData.get('description') as string).trim(),
      task_type: formData.get('task_type') as 'todo' | 'chore' | 'appointment',
      assigned_to: (formData.get('assigned_to') as string) || null,
      priority: parseInt(formData.get('priority') as string, 10),
    };

    // Validation
    let hasErrors = false;
    if (!data.title) {
      this.showError('title-error', 'Task title is required');
      hasErrors = true;
    }
    if (!data.task_type) {
      this.showError('task-type-error', 'Task type is required');
      hasErrors = true;
    }

    if (hasErrors) return;

    try {
      this.setSubmitting(true);
      await this.onSave(data, this.currentTask?.id);
      this.hide();
      this.resetForm();
    } catch (error) {
      this.showError('title-error', 'Failed to save task. Please try again.');
    } finally {
      this.setSubmitting(false);
    }
  }

  private clearErrors(): void {
    const errorElements = this.element.querySelectorAll('.form-error');
    errorElements.forEach(element => {
      element.textContent = '';
    });
  }

  private showError(elementId: string, message: string): void {
    const errorElement = this.element.querySelector(`#${elementId}`);
    if (errorElement) {
      errorElement.textContent = message;
    }
  }

  private setSubmitting(submitting: boolean): void {
    this.isSubmitting = submitting;
    const saveButton = this.element.querySelector('#save-button') as HTMLButtonElement;
    const buttonText = saveButton.querySelector('.button-text') as HTMLElement;
    const buttonLoading = saveButton.querySelector('.button-loading') as HTMLElement;

    if (submitting) {
      saveButton.disabled = true;
      buttonText.style.display = 'none';
      buttonLoading.style.display = 'inline';
    } else {
      saveButton.disabled = false;
      buttonText.style.display = 'inline';
      buttonLoading.style.display = 'none';
    }
  }

  private resetForm(): void {
    const form = this.element.querySelector('#task-form') as HTMLFormElement;
    form.reset();
    this.clearErrors();

    // Reset priority to default
    const prioritySelect = this.element.querySelector('#priority') as HTMLSelectElement;
    prioritySelect.value = '1';
  }

  public show(task?: Task): void {
    this.currentTask = task;

    // Update modal title and populate form if editing
    const titleElement = this.element.querySelector('#modal-title') as HTMLElement;
    const saveButton = this.element.querySelector('.button-text') as HTMLElement;

    if (task) {
      titleElement.textContent = 'Edit Task';
      saveButton.textContent = 'Update Task';
      this.populateForm(task);
    } else {
      titleElement.textContent = 'Add Task';
      saveButton.textContent = 'Save Task';
      this.resetForm();
    }

    // Show modal and backdrop
    this.backdrop.style.display = 'block';
    this.element.style.display = 'block';

    // Focus the title input
    const titleInput = this.element.querySelector('#task-title') as HTMLInputElement;
    setTimeout(() => titleInput.focus(), 100);

    // Add body class to prevent scrolling
    document.body.classList.add('modal-open');
  }

  public hide(): void {
    if (this.onCancel) {
      this.onCancel();
    }

    this.backdrop.style.display = 'none';
    this.element.style.display = 'none';
    this.currentTask = undefined;
    this.resetForm();

    // Remove body class to allow scrolling
    document.body.classList.remove('modal-open');
  }

  private populateForm(task: Task): void {
    (this.element.querySelector('#task-title') as HTMLInputElement).value = task.title;
    (this.element.querySelector('#task-description') as HTMLTextAreaElement).value =
      task.description;
    (this.element.querySelector('#task-type') as HTMLSelectElement).value = task.task_type;
    (this.element.querySelector('#assigned-to') as HTMLSelectElement).value =
      task.assigned_to || '';
    (this.element.querySelector('#priority') as HTMLSelectElement).value = task.priority.toString();
  }

  public destroy(): void {
    // Remove event listeners
    if (this.boundHandleClick) {
      this.element.removeEventListener('click', this.boundHandleClick);
    }
    if (this.boundHandleSubmit) {
      this.element.removeEventListener('submit', this.boundHandleSubmit);
    }

    // Remove elements from DOM
    if (this.element.parentNode) {
      this.element.parentNode.removeChild(this.element);
    }
    if (this.backdrop.parentNode) {
      this.backdrop.parentNode.removeChild(this.backdrop);
    }

    // Remove body class
    document.body.classList.remove('modal-open');
  }
}
