import { BasePage } from './base-page.js';
import { ComponentConfig } from '../common/types.js';
import { FamilyTasks } from '../tasks/family-tasks.js';
import { CalendarService } from '../calendar-events/calendar-events-service.js';
import { CalendarEvents } from '../calendar-events/calendar-events.js';

/**
 * DailyPage - Main daily view combining tasks and calendar
 * Replaces the old tasks page
 */
export class DailyPage extends BasePage {
  private familyTasks?: FamilyTasks;
  private calendarEvents?: CalendarEvents;
  private calendarService: CalendarService;
  private currentDate: Date = new Date();

  constructor(container: HTMLElement, config: ComponentConfig) {
    super(container, config, 'daily');
    this.calendarService = new CalendarService(config);
  }

  async init(): Promise<void> {
    try {
      this.showLoading('Loading daily view...');

      // Create the page structure
      this.container.innerHTML = this.renderPageContent();

      // Initialize components
      await this.initializeComponents();

      this.hideLoading();
    } catch (error) {
      this.showError('Failed to load daily view');
    }
  }

  private renderPageContent(): string {
    const dateFormatter = new Intl.DateTimeFormat('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });

    return `
      <div class="daily-page">
        <div class="daily-header">
          <div class="family-info">
            <div class="family-avatar">
              <span class="family-initial">S</span>
            </div>
            <div class="family-title">
              <h1>Our Family Hub  <button class="date-nav-btn" data-action="prev-day">←</button> ${dateFormatter.format(this.currentDate)} <button class="date-nav-btn" data-action="next-day">→</button> </h1>
            </div>
          </div>
        </div>
        <div class="daily-content">
          <div class="daily-left-panel">
            <div id="family-tasks-container"></div>
          </div>
          
          <div class="daily-right-panel">
            <div id="calendar-events-container"></div>
          </div>
        </div>
      </div>
    `;
  }

  private async initializeComponents(): Promise<void> {
    // Initialize family tasks component
    const tasksContainer = this.container.querySelector('#family-tasks-container') as HTMLElement;
    if (tasksContainer) {
      this.familyTasks = new FamilyTasks(tasksContainer, this.config);
    }

    // Initialize calendar events component
    const calendarEventsContainer = this.container.querySelector(
      '#calendar-events-container'
    ) as HTMLElement;
    if (calendarEventsContainer) {
      this.calendarEvents = new CalendarEvents(
        calendarEventsContainer,
        this.config,
        this.calendarService
      );
    }

    // Set up event listeners
    this.setupEventHandlers();
  }

  private setupEventHandlers(): void {
    this.container.addEventListener('click', this.handleClick.bind(this));

    // Listen for task-related events
    this.container.addEventListener(
      'add-task-requested',
      this.handleAddTaskRequest.bind(this) as EventListener
    );
  }

  private handleClick(e: Event): void {
    const target = e.target as HTMLElement;
    const action = target.getAttribute('data-action');

    switch (action) {
      case 'prev-day':
        this.navigateToDate(-1);
        break;
      case 'next-day':
        this.navigateToDate(1);
        break;
    }
  }

  private navigateToDate(dayOffset: number): void {
    const newDate = new Date(this.currentDate);
    newDate.setDate(newDate.getDate() + dayOffset);
    this.setDate(newDate);
  }

  private setDate(date: Date): void {
    this.currentDate = date;

    // Update header
    const dateFormatter = new Intl.DateTimeFormat('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });

    const titleElement = this.container.querySelector('.family-title h1');
    if (titleElement) {
      // Preserve the navigation buttons by updating innerHTML instead of textContent
      titleElement.innerHTML = `Our Family Hub  <button class="date-nav-btn" data-action="prev-day">←</button> ${dateFormatter.format(this.currentDate)} <button class="date-nav-btn" data-action="next-day">→</button>`;
    }

    // Update calendar events component
    if (this.calendarEvents) {
      this.calendarEvents.setDate(date);
    }

    // Update family tasks for the selected date
    if (this.familyTasks) {
      this.familyTasks.setDate(date);
    }
  }

  private handleAddTaskRequest(_: CustomEvent): void {
    // You could implement a task modal here or route to a different page
    // For now, let's just refresh the tasks
    if (this.familyTasks) {
      this.familyTasks.refresh();
    }
  }

  async refresh(): Promise<void> {
    if (this.familyTasks) {
      await this.familyTasks.refresh();
    }
    if (this.calendarEvents) {
      await this.calendarEvents.refresh();
    }
  }

  override destroy(): void {
    if (this.familyTasks) {
      this.familyTasks.destroy();
    }
    if (this.calendarEvents) {
      this.calendarEvents.destroy();
    }
    super.destroy();
  }
}
