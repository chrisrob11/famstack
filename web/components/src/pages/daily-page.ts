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

      // Add styles
      this.addStyles();

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
            <div class="family-details">
              <h1>Our Family Hub</h1>
              <div class="date-navigation">
                <button class="date-nav-btn" data-action="prev-day">←</button>
                <span class="current-date">${dateFormatter.format(this.currentDate)}</span>
                <button class="date-nav-btn" data-action="next-day">→</button>
              </div>
            </div>
          </div>
        </div>

        <div class="daily-content">
          <div class="daily-left-panel">
            <h2>Daily Chores</h2>
            <div id="family-tasks-container" class="daily-chores"></div>
          </div>

          <div class="daily-right-panel">
            <h2>Schedule</h2>
            <div id="calendar-events-container" class="daily-schedule"></div>
          </div>
        </div>
      </div>
    `;
  }

  private addStyles(): void {
    const styles = `
      <style id="daily-page-styles">
        .daily-page {
          max-width: 1400px;
          margin: 0 auto;
          padding: 2rem;
          background-color: #f9fafb;
          min-height: 100vh;
        }

        .daily-header {
          margin-bottom: 3rem;
        }

        .family-info {
          display: flex;
          align-items: center;
          gap: 1rem;
        }

        .family-avatar {
          width: 4rem;
          height: 4rem;
          background: linear-gradient(135deg, #8b5cf6, #a855f7);
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          color: white;
          font-size: 1.5rem;
          font-weight: bold;
          box-shadow: 0 4px 12px rgba(139, 92, 246, 0.3);
        }

        .family-details h1 {
          margin: 0 0 0.5rem 0;
          font-size: 2rem;
          font-weight: 700;
          color: #1f2937;
        }

        .date-navigation {
          display: flex;
          align-items: center;
          gap: 1rem;
        }

        .date-nav-btn {
          background: #f3f4f6;
          border: 1px solid #d1d5db;
          border-radius: 0.5rem;
          padding: 0.5rem 0.75rem;
          font-size: 1.125rem;
          cursor: pointer;
          transition: all 0.2s;
          color: #374151;
        }

        .date-nav-btn:hover {
          background: #e5e7eb;
          border-color: #9ca3af;
        }

        .current-date {
          font-size: 1.25rem;
          font-weight: 600;
          color: #374151;
          min-width: 200px;
          text-align: center;
        }

        .daily-content {
          display: grid;
          grid-template-columns: 60% 40%;
          gap: 3rem;
          align-items: start;
        }

        .daily-left-panel h2 {
          font-size: 1.75rem;
          font-weight: 700;
          color: #1f2937;
          margin: 0 0 2rem 0;
        }

        .daily-right-panel h2 {
          font-size: 1.75rem;
          font-weight: 700;
          color: #1f2937;
          margin: 0 0 2rem 0;
        }

        .daily-chores {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
          gap: 1.5rem;
        }

        .daily-schedule {
          background: white;
          border-radius: 1rem;
          padding: 1.5rem;
          box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
        }

        /* Family member task sections */
        .family-member-tasks {
          background: white;
          border-radius: 1rem;
          padding: 1.5rem;
          box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
        }

        .family-member-header {
          display: flex;
          align-items: center;
          gap: 0.75rem;
          margin-bottom: 1rem;
          padding-bottom: 0.75rem;
          border-bottom: 1px solid #e5e7eb;
        }

        .member-avatar {
          width: 2.5rem;
          height: 2.5rem;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          color: white;
          font-weight: 600;
          font-size: 0.875rem;
        }

        .member-name {
          font-size: 1.125rem;
          font-weight: 600;
          color: #1f2937;
          margin: 0;
        }

        .add-task-btn {
          margin-left: auto;
          background: none;
          border: none;
          color: #8b5cf6;
          font-size: 1.25rem;
          cursor: pointer;
          padding: 0.25rem;
          border-radius: 0.25rem;
          transition: background-color 0.2s;
        }

        .add-task-btn:hover {
          background: #f3f0ff;
        }

        .task-list {
          display: flex;
          flex-direction: column;
          gap: 0.75rem;
        }

        .task-item {
          display: flex;
          align-items: center;
          gap: 0.75rem;
          padding: 0.5rem 0;
        }

        .task-checkbox {
          width: 1.25rem;
          height: 1.25rem;
          border: 2px solid #d1d5db;
          border-radius: 0.25rem;
          cursor: pointer;
          transition: all 0.2s;
        }

        .task-checkbox:checked {
          background: #8b5cf6;
          border-color: #8b5cf6;
        }

        .task-label {
          font-size: 1rem;
          color: #374151;
          cursor: pointer;
          flex: 1;
        }

        .task-label.completed {
          text-decoration: line-through;
          color: #9ca3af;
        }

        /* Schedule time slots */
        .time-slot {
          display: flex;
          align-items: center;
          padding: 0.75rem 0;
          border-bottom: 1px solid #f3f4f6;
          font-size: 0.875rem;
          color: #6b7280;
        }

        .time-slot:last-child {
          border-bottom: none;
        }

        /* Responsive design */
        @media (max-width: 1200px) {
          .daily-chores {
            grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
            gap: 1rem;
          }
        }

        @media (max-width: 1024px) {
          .daily-content {
            grid-template-columns: 1fr;
            gap: 2rem;
          }

          .daily-chores {
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 1.5rem;
          }

          .daily-page {
            padding: 1rem;
          }
        }

        @media (max-width: 768px) {
          .daily-chores {
            grid-template-columns: 1fr;
            gap: 1rem;
          }
        }

        @media (max-width: 640px) {
          .family-info {
            flex-direction: column;
            align-items: flex-start;
            gap: 1rem;
          }

          .date-navigation {
            justify-content: center;
            width: 100%;
          }
        }
      </style>
    `;

    // Remove existing styles
    const existingStyles = document.getElementById('daily-page-styles');
    if (existingStyles) {
      existingStyles.remove();
    }

    // Add styles to head
    document.head.insertAdjacentHTML('beforeend', styles);
  }

  private async initializeComponents(): Promise<void> {
    // Initialize family tasks component (left side - Daily Chores)
    const tasksContainer = this.container.querySelector('#family-tasks-container') as HTMLElement;
    if (tasksContainer) {
      this.familyTasks = new FamilyTasks(tasksContainer, this.config);
    }

    // Initialize calendar events component (right side - Today's Schedule)
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

    // Update header date
    const dateFormatter = new Intl.DateTimeFormat('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });

    const currentDateElement = this.container.querySelector('.current-date');
    if (currentDateElement) {
      currentDateElement.textContent = dateFormatter.format(this.currentDate);
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

    // Remove styles
    const styles = document.getElementById('daily-page-styles');
    if (styles) {
      styles.remove();
    }

    super.destroy();
  }
}

export default DailyPage;
