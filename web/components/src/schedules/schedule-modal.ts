import { ComponentConfig } from '../common/types.js';
import { TaskSchedule, CreateScheduleRequest } from './schedule-service.js';

export type ScheduleFormData = Omit<CreateScheduleRequest, 'family_id'>;

export interface ScheduleModalOptions {
  onSave: (data: ScheduleFormData, scheduleId?: string) => Promise<void>;
  onCancel?: () => void;
}

export class ScheduleModal {
  private element!: HTMLElement;
  private backdrop!: HTMLElement;
  private currentSchedule: TaskSchedule | undefined;
  private isSubmitting: boolean = false;
  private onSave: (data: ScheduleFormData, scheduleId?: string) => Promise<void>;
  private onCancel: (() => void) | undefined;
  private boundHandleClick?: (e: Event) => void;
  private boundHandleSubmit?: (e: Event) => void;

  constructor(container: HTMLElement, _config: ComponentConfig, options: ScheduleModalOptions) {
    this.onSave = options.onSave;
    this.onCancel = options.onCancel;

    this.createElement(container);
    this.attachEventListeners();
    this.populateUserOptions();
  }

  private createElement(container: HTMLElement): void {
    // Create modal
    this.element = document.createElement('div');
    this.element.className = 'modal';
    this.element.id = 'schedule-modal';
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
          <h2 id="modal-title">Create New Schedule</h2>
          <button class="modal-close" data-action="close-modal" type="button">×</button>
        </div>
        <form id="schedule-form">
          <input type="hidden" id="schedule-id" name="id">
          
          <div class="form-group">
            <label for="schedule-title">Title *</label>
            <input type="text" id="schedule-title" name="title" required 
                   placeholder="e.g., Take out trash, Do homework">
          </div>

          <div class="form-group">
            <label for="schedule-description">Description</label>
            <textarea id="schedule-description" name="description" rows="3"
                      placeholder="Optional description..."></textarea>
          </div>

          <div class="form-row-three">
            <div class="form-group">
              <label for="schedule-type">Type *</label>
              <select id="schedule-type" name="task_type" required>
                <option value="chore">Chore</option>
                <option value="todo">To-Do</option>
                <option value="appointment">Appointment</option>
              </select>
            </div>

            <div class="form-group">
              <label for="schedule-assigned-to">Assigned To</label>
              <select id="schedule-assigned-to" name="assigned_to">
                <option value="">Unassigned</option>
              </select>
            </div>

            <div class="form-group">
              <label for="schedule-priority">Priority</label>
              <select id="schedule-priority" name="priority">
                <option value="0">Low</option>
                <option value="1">Normal</option>
                <option value="2">High</option>
                <option value="3">Urgent</option>
              </select>
            </div>
          </div>

          <div class="form-group">
            <label>Days of the Week *</label>
            <div class="days-of-week">
              <label class="day-checkbox">
                <input type="checkbox" name="days_of_week" value="monday"> Mon
              </label>
              <label class="day-checkbox">
                <input type="checkbox" name="days_of_week" value="tuesday"> Tue
              </label>
              <label class="day-checkbox">
                <input type="checkbox" name="days_of_week" value="wednesday"> Wed
              </label>
              <label class="day-checkbox">
                <input type="checkbox" name="days_of_week" value="thursday"> Thu
              </label>
              <label class="day-checkbox">
                <input type="checkbox" name="days_of_week" value="friday"> Fri
              </label>
              <label class="day-checkbox">
                <input type="checkbox" name="days_of_week" value="saturday"> Sat
              </label>
              <label class="day-checkbox">
                <input type="checkbox" name="days_of_week" value="sunday"> Sun
              </label>
            </div>
          </div>


          <div class="modal-actions">
            <button type="button" class="btn btn-secondary" data-action="close-modal">Cancel</button>
            <button type="submit" class="btn btn-primary" id="submit-btn">Create Schedule</button>
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
  }

  private async populateUserOptions(): Promise<void> {
    const select = this.element.querySelector('#schedule-assigned-to') as HTMLSelectElement;
    if (select) {
      // TODO: Replace with actual user fetching
      select.innerHTML = `
        <option value="">Unassigned</option>
        <option value="user1">John Smith</option>
        <option value="user2">Jane Smith</option>
        <option value="user3">Bobby Smith</option>
      `;
    }
  }

  public showCreate(): void {
    this.currentSchedule = undefined;

    const title = this.element.querySelector('#modal-title') as HTMLElement;
    const submitBtn = this.element.querySelector('#submit-btn') as HTMLButtonElement;

    title.textContent = 'Create New Schedule';
    submitBtn.textContent = 'Create Schedule';

    this.resetForm();
    this.show();
  }

  public showEdit(schedule: TaskSchedule): void {
    this.currentSchedule = schedule;

    const title = this.element.querySelector('#modal-title') as HTMLElement;
    const submitBtn = this.element.querySelector('#submit-btn') as HTMLButtonElement;

    title.textContent = 'Edit Schedule';
    submitBtn.textContent = 'Update Schedule';

    this.populateForm(schedule);
    this.show();
  }

  private show(): void {
    this.element.style.display = 'flex';
    this.backdrop.style.display = 'block';

    const titleInput = this.element.querySelector('#schedule-title') as HTMLInputElement;
    titleInput?.focus();
  }

  public hide(): void {
    this.element.style.display = 'none';
    this.backdrop.style.display = 'none';
    this.resetForm();
  }

  private resetForm(): void {
    const form = this.element.querySelector('#schedule-form') as HTMLFormElement;
    form.reset();

    // Set default priority
    const prioritySelect = form.querySelector('#schedule-priority') as HTMLSelectElement;
    prioritySelect.value = '1';

    // Clear any error messages
    const errors = form.querySelectorAll('.field-error');
    errors.forEach(error => error.remove());
  }

  private populateForm(schedule: TaskSchedule): void {
    const form = this.element.querySelector('#schedule-form') as HTMLFormElement;

    (form.querySelector('#schedule-id') as HTMLInputElement).value = schedule.id;
    (form.querySelector('#schedule-title') as HTMLInputElement).value = schedule.title;
    (form.querySelector('#schedule-description') as HTMLTextAreaElement).value =
      schedule.description || '';
    (form.querySelector('#schedule-type') as HTMLSelectElement).value = schedule.task_type;
    (form.querySelector('#schedule-assigned-to') as HTMLSelectElement).value =
      schedule.assigned_to || '';
    (form.querySelector('#schedule-priority') as HTMLSelectElement).value =
      schedule.priority.toString();

    // Set checkboxes for days of the week
    const dayCheckboxes = form.querySelectorAll(
      'input[name="days_of_week"]'
    ) as NodeListOf<HTMLInputElement>;
    dayCheckboxes.forEach(checkbox => {
      checkbox.checked = schedule.days_of_week.includes(checkbox.value);
    });
  }

  private handleClick(e: Event): void {
    const target = e.target as HTMLElement;
    const action = target.getAttribute('data-action');

    if (action === 'close-modal') {
      e.preventDefault();
      this.hide();
      this.onCancel?.();
    }
  }

  private handleSubmit(e: Event): void {
    e.preventDefault();
    if (this.isSubmitting) return;

    const form = e.target as HTMLFormElement;
    this.handleFormSubmit(form);
  }

  private async handleFormSubmit(form: HTMLFormElement): Promise<void> {
    if (this.isSubmitting) return;

    this.isSubmitting = true;

    try {
      const formData = this.getFormData(form);
      const scheduleId = this.currentSchedule?.id;

      await this.onSave(formData, scheduleId);
      this.hide();
    } catch (error) {
      this.showFormError(form, error instanceof Error ? error.message : 'Failed to save schedule');
    } finally {
      this.isSubmitting = false;
    }
  }

  private getFormData(form: HTMLFormElement): ScheduleFormData {
    const formData = new FormData(form);

    // Get selected days
    const selectedDays: string[] = [];
    const dayCheckboxes = form.querySelectorAll(
      'input[name="days_of_week"]:checked'
    ) as NodeListOf<HTMLInputElement>;
    dayCheckboxes.forEach(checkbox => {
      selectedDays.push(checkbox.value);
    });

    if (selectedDays.length === 0) {
      throw new Error('Please select at least one day of the week');
    }

    return {
      title: formData.get('title') as string,
      description: (formData.get('description') as string) || '',
      task_type: formData.get('task_type') as 'todo' | 'chore' | 'appointment',
      assigned_to: (formData.get('assigned_to') as string) || null,
      days_of_week: selectedDays,
      time_of_day: null,
      priority: parseInt(formData.get('priority') as string),
      points: 0,
    };
  }

  private showFormError(form: HTMLFormElement, message: string): void {
    // Clear existing errors
    const existingErrors = form.querySelectorAll('.field-error');
    existingErrors.forEach(error => error.remove());

    // Add new error
    const errorDiv = document.createElement('div');
    errorDiv.className = 'field-error form-error-general';
    errorDiv.textContent = message;
    errorDiv.style.cssText = `
      color: #e53e3e;
      font-size: 0.875rem;
      margin-top: 0.5rem;
      padding: 0.5rem;
      background: #fed7d7;
      border-radius: 4px;
    `;

    form.appendChild(errorDiv);
  }

  public destroy(): void {
    if (this.boundHandleClick) {
      this.element.removeEventListener('click', this.boundHandleClick);
    }
    if (this.boundHandleSubmit) {
      this.element.removeEventListener('submit', this.boundHandleSubmit);
    }

    this.element.remove();
    this.backdrop.remove();
  }
}
