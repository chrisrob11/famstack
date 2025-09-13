import { ComponentConfig } from '../common/types.js';
import { TaskService, TasksResponse } from './task-service.js';
import { PersonTasks } from './person-tasks.js';
import { TaskModal, TaskFormData } from './task-modal.js';
import { CreateTaskData } from './task-service.js';

/**
 * FamilyTasks - Component for managing and displaying tasks for all family members
 * Creates a flexible grid layout with PersonTasks components
 */
export class FamilyTasks {
  private container: HTMLElement;
  private config: ComponentConfig;
  private taskService: TaskService;
  private tasks: TasksResponse | null = null;
  private personTaskComponents: Map<string, PersonTasks> = new Map();
  private taskModal?: TaskModal;
  private boundHandleClick?: (e: Event) => void;
  private boundHandleTaskToggle?: (e: CustomEvent) => void;

  constructor(container: HTMLElement, config: ComponentConfig) {
    this.container = container;
    this.config = config;
    this.taskService = new TaskService(config);
    this.init();
  }

  private init(): void {
    this.setupEventListeners();
    this.loadTasks();
  }

  private setupEventListeners(): void {
    this.boundHandleClick = this.handleClick.bind(this);
    this.boundHandleTaskToggle = this.handleTaskToggle.bind(this);

    this.container.addEventListener('click', this.boundHandleClick);
    this.container.addEventListener('task-toggle', this.boundHandleTaskToggle as any);
    this.container.addEventListener('person-add-task', this.handlePersonAddTask.bind(this) as any);
    this.container.addEventListener('task-update', this.handleTaskUpdate.bind(this) as any);
  }

  private async loadTasks(): Promise<void> {
    try {
      this.renderLoading();
      const tasksData = await this.taskService.getTasks();
      this.tasks = tasksData;
      this.renderFamilyTasks();
    } catch (error) {
      this.renderError('Failed to load tasks');
    }
  }

  private renderLoading(): void {
    this.container.innerHTML = `
      <div class="family-tasks-loading">
        <div class="loading-spinner"></div>
        <p>Loading tasks...</p>
      </div>
    `;
  }

  private renderError(message: string): void {
    this.container.innerHTML = `
      <div class="family-tasks-error">
        <p class="error-message">${message}</p>
        <button class="retry-btn" data-action="retry">Try Again</button>
      </div>
    `;
  }

  private renderFamilyTasks(): void {
    if (!this.tasks) {
      this.renderError('No tasks found');
      return;
    }

    // Clean up existing person task components
    this.personTaskComponents.forEach(component => component.destroy());
    this.personTaskComponents.clear();

    this.container.innerHTML = `
      <div class="family-tasks">
        <div class="family-tasks-header">
          <h2>Daily Chores</h2>
        </div>
        <div class="family-tasks-grid">
          ${this.renderPersonTaskContainers()}
        </div>
      </div>
    `;

    // Initialize PersonTasks components for each family member
    this.initializePersonTasks();

    // Initialize task modal
    this.initializeTaskModal();
  }

  private renderPersonTaskContainers(): string {
    if (!this.tasks) return '';

    const userColumns = Object.values(this.tasks.tasks_by_user);
    const containers: string[] = [];

    userColumns.forEach(column => {
      const user = column.user;
      if (user.name !== 'Unassigned') {
        containers.push(`
          <div class="person-tasks-container" data-user-id="${user.id}">
            <!-- PersonTasks component will be initialized here -->
          </div>
        `);
      }
    });

    return containers.join('');
  }

  private initializePersonTasks(): void {
    if (!this.tasks) return;

    const userColumns = Object.values(this.tasks.tasks_by_user);

    userColumns.forEach(column => {
      const user = column.user;
      if (user.name !== 'Unassigned') {
        const container = this.container.querySelector(
          `[data-user-id="${user.id}"]`
        ) as HTMLElement;
        if (container) {
          const personTasks = new PersonTasks(container, this.config, user);
          personTasks.setTasks(column.tasks || []);
          this.personTaskComponents.set(user.id, personTasks);
        }
      }
    });
  }

  private initializeTaskModal(): void {
    if (this.taskModal) {
      this.taskModal.destroy();
    }

    // Get family members for assignment dropdown
    const familyMembers = this.tasks
      ? Object.values(this.tasks.tasks_by_user)
          .map(column => column.user)
          .filter(user => user.name !== 'Unassigned')
      : [];

    this.taskModal = new TaskModal(document.body, this.config, {
      onSave: this.handleSaveTask.bind(this),
      familyMembers,
    });
  }

  private handleClick(e: Event): void {
    const target = e.target as HTMLElement;
    const action = target.getAttribute('data-action');

    switch (action) {
      case 'add-task':
        this.handleAddTask();
        break;
      case 'retry':
        this.loadTasks();
        break;
    }
  }

  private handleAddTask(): void {
    if (this.taskModal) {
      this.taskModal.show();
    }
  }

  private async handleSaveTask(data: TaskFormData, taskId?: string): Promise<void> {
    if (taskId) {
      // Update existing task
      await this.taskService.updateTask(taskId, data);
    } else {
      // Create new task - add family_id
      const createData: CreateTaskData = {
        ...data,
        assigned_to: data.assigned_to || undefined,
        family_id: 'fam1', // TODO: Get actual family ID
      };
      await this.taskService.createTask(createData);
    }

    // Refresh the tasks display
    await this.loadTasks();
  }

  private async handleTaskToggle(e: CustomEvent): Promise<void> {
    const { taskId, newStatus, originalState, checkbox } = e.detail;

    try {
      await this.taskService.updateTask(taskId, { status: newStatus });

      // Update local state in the person component
      for (const personComponent of this.personTaskComponents.values()) {
        personComponent.updateTaskStatus(taskId, newStatus);
      }
    } catch (error) {
      // Revert checkbox state on error
      checkbox.checked = originalState;
    }
  }

  private async handlePersonAddTask(e: CustomEvent): Promise<void> {
    const { userId } = e.detail;

    // Create a simple new task for this person
    const createData: CreateTaskData = {
      title: 'New Task',
      description: '',
      task_type: 'todo',
      assigned_to: userId,
      family_id: 'fam1', // TODO: Get actual family ID
    };

    await this.taskService.createTask(createData);

    // Refresh the tasks display
    await this.loadTasks();
  }

  private async handleTaskUpdate(e: CustomEvent): Promise<void> {
    const { taskId, title } = e.detail;

    try {
      await this.taskService.updateTask(taskId, { title });

      // Update local state in the person components
      for (const personComponent of this.personTaskComponents.values()) {
        const task = personComponent['tasks']?.find((t: any) => t.id === taskId);
        if (task) {
          task.title = title;
        }
      }
    } catch (error) {
      // Refresh to revert changes on error
      await this.loadTasks();
    }
  }

  public async refresh(): Promise<void> {
    await this.loadTasks();
  }

  public destroy(): void {
    // Clean up person task components
    this.personTaskComponents.forEach(component => component.destroy());
    this.personTaskComponents.clear();

    // Remove event listeners
    if (this.boundHandleClick) {
      this.container.removeEventListener('click', this.boundHandleClick);
    }
    if (this.boundHandleTaskToggle) {
      this.container.removeEventListener('task-toggle', this.boundHandleTaskToggle as any);
    }

    // Clean up task modal
    if (this.taskModal) {
      this.taskModal.destroy();
    }
  }
}
