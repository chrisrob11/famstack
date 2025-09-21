import { ComponentConfig } from '../common/types.js';
import { TaskSchedule, parseDaysOfWeek } from './schedule-service.js';

export class ScheduleCard {
  private schedule: TaskSchedule;
  private element: HTMLElement;
  private onScheduleUpdate?: (schedule: TaskSchedule) => void;
  private onScheduleDelete?: (scheduleId: string) => void;
  private onScheduleToggle?: (scheduleId: string) => void;
  private onScheduleEdit?: (schedule: TaskSchedule) => void;
  private boundHandleClick?: (e: Event) => void;

  constructor(
    schedule: TaskSchedule,
    _config: ComponentConfig,
    options: {
      onScheduleUpdate?: (schedule: TaskSchedule) => void;
      onScheduleDelete?: (scheduleId: string) => void;
      onScheduleToggle?: (scheduleId: string) => void;
      onScheduleEdit?: (schedule: TaskSchedule) => void;
    } = {}
  ) {
    this.schedule = schedule;
    if (options.onScheduleUpdate) {
      this.onScheduleUpdate = options.onScheduleUpdate;
    }
    if (options.onScheduleDelete) {
      this.onScheduleDelete = options.onScheduleDelete;
    }
    if (options.onScheduleToggle) {
      this.onScheduleToggle = options.onScheduleToggle;
    }
    if (options.onScheduleEdit) {
      this.onScheduleEdit = options.onScheduleEdit;
    }
    this.element = this.createElement();
    this.attachEventListeners();
  }

  private createElement(): HTMLElement {
    const cardElement = document.createElement('div');
    cardElement.className = `schedule-card ${this.schedule.active ? 'active' : 'inactive'}`;
    cardElement.setAttribute('data-schedule-id', this.schedule.id);
    cardElement.innerHTML = this.getCardHTML();
    return cardElement;
  }

  private getCardHTML(): string {
    const days = parseDaysOfWeek(this.schedule.days_of_week);
    const daysString = days.map(day => day.charAt(0).toUpperCase() + day.slice(1, 3)).join(', ');

    const timeDisplay = this.schedule.time_of_day
      ? ` at ${this.formatTime(this.schedule.time_of_day)}`
      : '';

    return `
      <div class="schedule-header">
        <div class="schedule-status ${this.schedule.active ? 'active' : 'inactive'}">
          ${this.schedule.active ? 'Active' : 'Inactive'}
        </div>
        <div class="schedule-title">
          ${this.schedule.title}
        </div>
        <div class="schedule-actions">
          <button class="schedule-action-btn" data-action="toggle" 
                  style="color: ${this.schedule.active ? 'orange' : 'green'};" 
                  title="${this.schedule.active ? 'Deactivate' : 'Activate'}">
            ${this.schedule.active ? '⏸' : '▶'}
          </button>
          <button class="schedule-action-btn" data-action="edit" style="color: blue;" title="Edit">
            ✎
          </button>
          <button class="schedule-action-btn" data-action="delete" style="color: red;" title="Delete">
            ×
          </button>
        </div>
      </div>
      ${
        this.schedule.description
          ? `
        <div class="schedule-description">
          ${this.schedule.description}
        </div>
      `
          : ''
      }
      <div class="schedule-details">
        <div class="schedule-recurrence">
          <span class="schedule-days">${daysString}${timeDisplay}</span>
        </div>
        <div class="schedule-meta">
          <span class="schedule-type ${this.schedule.task_type}">${this.schedule.task_type}</span>
          ${this.schedule.assigned_to ? `<span class="schedule-assignee">→ ${this.schedule.assigned_to}</span>` : '<span class="schedule-assignee">→ Unassigned</span>'}
          ${this.schedule.points > 0 ? `<span class="schedule-points">${this.schedule.points} pts</span>` : ''}
        </div>
      </div>
    `;
  }

  private formatTime(time: string): string {
    // Convert 24-hour time to 12-hour format
    const parts = time.split(':').map(Number);
    if (parts.length < 2) return time; // Return original if invalid format

    const hours = parts[0];
    const minutes = parts[1];
    if (hours === undefined || minutes === undefined) return time;

    const period = hours >= 12 ? 'PM' : 'AM';
    const displayHours = hours === 0 ? 12 : hours > 12 ? hours - 12 : hours;
    return `${displayHours}:${minutes.toString().padStart(2, '0')} ${period}`;
  }

  private attachEventListeners(): void {
    this.boundHandleClick = this.handleClick.bind(this);
    this.element.addEventListener('click', this.boundHandleClick);
  }

  private handleClick(e: Event): void {
    const target = e.target as HTMLElement;
    const action = target.getAttribute('data-action');

    if (action) {
      e.preventDefault();
      e.stopPropagation();
      this.handleAction(action);
    }
  }

  private async handleAction(action: string): Promise<void> {
    switch (action) {
      case 'toggle':
        this.onScheduleToggle?.(this.schedule.id);
        break;
      case 'edit':
        this.onScheduleEdit?.(this.schedule);
        break;
      case 'delete':
        if (
          confirm(
            'Are you sure you want to delete this schedule? This will not affect tasks already created from it.'
          )
        ) {
          this.onScheduleDelete?.(this.schedule.id);
        }
        break;
    }
  }

  public updateSchedule(updatedSchedule: TaskSchedule): void {
    this.schedule = updatedSchedule;
    this.element.innerHTML = this.getCardHTML();
    this.element.className = `schedule-card ${this.schedule.active ? 'active' : 'inactive'}`;
    this.onScheduleUpdate?.(this.schedule);
  }

  public getElement(): HTMLElement {
    return this.element;
  }

  public getSchedule(): TaskSchedule {
    return this.schedule;
  }

  public destroy(): void {
    if (this.boundHandleClick) {
      this.element.removeEventListener('click', this.boundHandleClick);
    }
    this.element.remove();
  }
}
