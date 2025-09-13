import { ComponentConfig } from '../common/types.js';
import { TaskSchedule, ScheduleService } from './schedule-service.js';
import { ScheduleCard } from './schedule-card.js';
import { ScheduleModal, ScheduleFormData } from './schedule-modal.js';
import { ComponentUtils } from '../common/component-utils.js';

export class ScheduleList {
  private config: ComponentConfig;
  private container: HTMLElement;
  private scheduleService: ScheduleService;
  private scheduleCards: Map<string, ScheduleCard> = new Map();
  private scheduleModal!: ScheduleModal;
  private boundHandleClick?: (e: Event) => void;

  constructor(container: HTMLElement, config: ComponentConfig) {
    this.container = container;
    this.config = config;
    this.scheduleService = new ScheduleService(config);
    this.init();
  }

  private init(): void {
    this.container.className = 'schedule-list-container';
    this.render();
    this.setupEventListeners();
    this.initializeModal();
    this.loadSchedules();
  }

  private render(): void {
    this.container.innerHTML = `
      <div class="schedule-header">
        <h1>Task Schedules</h1>
        <button class="btn btn-primary" data-action="add-schedule">
          + Add Schedule
        </button>
      </div>

      <div class="schedule-grid" id="schedule-grid">
        <div class="loading">Loading schedules...</div>
      </div>
    `;
  }

  private setupEventListeners(): void {
    this.boundHandleClick = this.handleClick.bind(this);
    this.container.addEventListener('click', this.boundHandleClick);
  }

  private initializeModal(): void {
    this.scheduleModal = new ScheduleModal(this.container, this.config, {
      onSave: this.handleModalSave.bind(this),
      onCancel: () => {},
    });
  }

  private async handleModalSave(data: ScheduleFormData, scheduleId?: string): Promise<void> {
    if (scheduleId) {
      // Edit existing schedule
      const updatedSchedule = await this.scheduleService.updateSchedule(scheduleId, data);
      const scheduleCard = this.scheduleCards.get(scheduleId);
      if (scheduleCard) {
        scheduleCard.updateSchedule(updatedSchedule);
      }
      this.showSuccess('Schedule updated successfully!');
    } else {
      // Create new schedule
      const scheduleData = { ...data, family_id: 'fam1' };
      const newSchedule = await this.scheduleService.createSchedule(scheduleData);
      this.addScheduleToList(newSchedule);
      this.showSuccess('Schedule created successfully!');
    }
  }

  private async loadSchedules(): Promise<void> {
    try {
      const schedules = await this.scheduleService.listSchedules();
      this.renderSchedules(schedules);
    } catch (error) {
      this.showError('Failed to load schedules');
    }
  }

  private renderSchedules(schedules: TaskSchedule[]): void {
    const grid = this.container.querySelector('#schedule-grid') as HTMLElement;

    if (schedules.length === 0) {
      grid.innerHTML = `
        <div class="empty-state">
          <p>No schedules created yet.</p>
          <p>Create your first schedule to automatically generate recurring tasks.</p>
        </div>
      `;
      return;
    }

    grid.innerHTML = '';

    // Clear existing cards
    this.scheduleCards.forEach(card => card.destroy());
    this.scheduleCards.clear();

    // Create new cards
    schedules.forEach(schedule => {
      const scheduleCard = new ScheduleCard(schedule, this.config, {
        onScheduleUpdate: this.handleScheduleUpdate.bind(this),
        onScheduleDelete: this.handleScheduleDelete.bind(this),
        onScheduleToggle: this.handleScheduleToggle.bind(this),
        onScheduleEdit: this.handleScheduleEdit.bind(this),
      });

      grid.appendChild(scheduleCard.getElement());
      this.scheduleCards.set(schedule.id, scheduleCard);
    });
  }

  private addScheduleToList(schedule: TaskSchedule): void {
    const grid = this.container.querySelector('#schedule-grid') as HTMLElement;
    const emptyState = grid.querySelector('.empty-state');

    if (emptyState) {
      grid.innerHTML = '';
    }

    const scheduleCard = new ScheduleCard(schedule, this.config, {
      onScheduleUpdate: this.handleScheduleUpdate.bind(this),
      onScheduleDelete: this.handleScheduleDelete.bind(this),
      onScheduleToggle: this.handleScheduleToggle.bind(this),
      onScheduleEdit: this.handleScheduleEdit.bind(this),
    });

    grid.appendChild(scheduleCard.getElement());
    this.scheduleCards.set(schedule.id, scheduleCard);
  }

  private handleClick(e: Event): void {
    const target = e.target as HTMLElement;
    const action = target.getAttribute('data-action');

    switch (action) {
      case 'add-schedule':
        this.scheduleModal.showCreate();
        break;
    }
  }

  private handleScheduleEdit(schedule: TaskSchedule): void {
    this.scheduleModal.showEdit(schedule);
  }

  private async handleScheduleToggle(scheduleId: string): Promise<void> {
    try {
      const result = await this.scheduleService.toggleSchedule(scheduleId);
      const scheduleCard = this.scheduleCards.get(scheduleId);
      if (scheduleCard) {
        const updatedSchedule = { ...scheduleCard.getSchedule(), active: result.active };
        scheduleCard.updateSchedule(updatedSchedule);
      }
      this.showSuccess(`Schedule ${result.active ? 'activated' : 'deactivated'}`);
    } catch (error) {
      this.showError('Failed to toggle schedule');
    }
  }

  private async handleScheduleDelete(scheduleId: string): Promise<void> {
    try {
      await this.scheduleService.deleteSchedule(scheduleId);

      const scheduleCard = this.scheduleCards.get(scheduleId);
      if (scheduleCard) {
        scheduleCard.destroy();
        this.scheduleCards.delete(scheduleId);
      }

      // Show empty state if no schedules left
      if (this.scheduleCards.size === 0) {
        const grid = this.container.querySelector('#schedule-grid') as HTMLElement;
        grid.innerHTML = `
          <div class="empty-state">
            <p>No schedules created yet.</p>
            <p>Create your first schedule to automatically generate recurring tasks.</p>
          </div>
        `;
      }

      this.showSuccess('Schedule deleted successfully');
    } catch (error) {
      this.showError('Failed to delete schedule');
    }
  }

  private handleScheduleUpdate(_: TaskSchedule): void {
    // Handle any additional schedule update logic if needed
  }

  private showError(message: string): void {
    ComponentUtils.showError(message);
  }

  private showSuccess(message: string): void {
    ComponentUtils.showSuccess(message);
  }

  public async refresh(): Promise<void> {
    await this.loadSchedules();
  }

  public destroy(): void {
    // Clean up schedule cards
    this.scheduleCards.forEach(card => card.destroy());
    this.scheduleCards.clear();

    // Clean up modal
    if (this.scheduleModal) {
      this.scheduleModal.destroy();
    }

    // Clean up event listeners
    if (this.boundHandleClick) {
      this.container.removeEventListener('click', this.boundHandleClick);
    }
  }
}
