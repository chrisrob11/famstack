import { ComponentConfig } from '../common/types.js';
import { TaskCard, Task } from './task-card.js';
import { TaskService, TasksResponse } from './task-service.js';
import { TaskListRenderer } from './task-list-renderer.js';
import { TaskDragDropManager } from './task-drag-drop-manager.js';
import { ComponentUtils } from '../common/component-utils.js';
import { stateManager, AppState } from './state-manager.js';

/**
 * TaskList - Coordinating component that uses composition pattern
 * Single responsibility: Orchestrate between renderer, drag-drop, service, and cards
 */
export class TaskList {
  private config: ComponentConfig;
  private container: HTMLElement;
  public taskCards: Map<string, TaskCard> = new Map();
  private boundHandleClick?: (e: Event) => void;
  private boundHandleSubmit?: (e: Event) => void;
  private isSubmitting: boolean = false;
  private unsubscribeFromState?: () => void;
  
  // Separated responsibilities using composition
  private taskService: TaskService;
  private renderer: TaskListRenderer;
  private dragDropManager: TaskDragDropManager;

  constructor(container: HTMLElement, config: ComponentConfig) {
    this.container = container;
    this.config = config;
    this.taskService = new TaskService(config);
    this.renderer = new TaskListRenderer(container);
    this.dragDropManager = new TaskDragDropManager(container, {
      onTaskReorder: this.handleTaskReorder.bind(this)
    });
    this.init();
  }

  private init(): void {
    this.container.className = 'task-list-container';
    
    // Setup event listeners once during initialization
    this.setupEventListeners();
    
    // Subscribe to state changes
    this.unsubscribeFromState = stateManager.subscribe(this.handleStateChange.bind(this));
    
    // Load initial data
    this.loadTasks();
  }

  private setupEventListeners(): void {
    // Clean up any existing listeners
    if (this.boundHandleClick) {
      this.container.removeEventListener('click', this.boundHandleClick);
    }
    if (this.boundHandleSubmit) {
      this.container.removeEventListener('submit', this.boundHandleSubmit);
    }

    // Bind and attach new listeners
    this.boundHandleClick = this.handleClick.bind(this);
    this.boundHandleSubmit = this.handleSubmit.bind(this);
    this.container.addEventListener('click', this.boundHandleClick);
    this.container.addEventListener('submit', this.boundHandleSubmit);
  }

  private handleStateChange(state: AppState): void {
    if (state.loading) {
      this.renderer.renderLoading();
    } else if (state.error) {
      this.renderer.renderError(state.error);
    } else if (state.tasks) {
      this.renderTasks(state.tasks);
    }
  }

  private async loadTasks(): Promise<void> {
    try {
      stateManager.setLoading(true);
      const tasksData = await this.taskService.getTasks();
      stateManager.setTasks(tasksData);
    } catch (error) {
      stateManager.setError('Failed to load tasks');
    }
  }

  private renderTasks(tasksData: TasksResponse): void {
    // Use renderer for HTML generation
    this.renderer.renderTaskList(tasksData);

    // Create task cards and set up drag/drop
    this.createTaskCards();
    this.dragDropManager.setupSortable();
  }


  private createTaskCards(): void {
    const state = stateManager.getState();
    if (!state.tasks) return;

    // Clear existing cards
    this.taskCards.forEach(card => card.destroy());
    this.taskCards.clear();

    Object.values(state.tasks.tasks_by_user).forEach(column => {
      // Handle null tasks from API
      const tasks = column.tasks || [];
      
      tasks.forEach(task => {
        const container = this.container.querySelector(`[data-task-container="${task.id}"]`) as HTMLElement;
        if (container) {
          const taskCard = new TaskCard(task, this.config, {
            onTaskUpdate: this.handleTaskUpdate.bind(this),
            onTaskDelete: this.handleTaskDelete.bind(this)
          });
          container.appendChild(taskCard.getElement());
          this.taskCards.set(task.id, taskCard);
        }
      });
    });
  }

  private handleClick(e: Event): void {
    const target = e.target as HTMLElement;
    const action = target.getAttribute('data-action');
    
    switch (action) {
      case 'add-task':
        this.showAddTaskModal();
        break;
      case 'close-modal':
        this.hideAddTaskModal();
        break;
    }
  }

  private handleSubmit(e: Event): void {
    e.preventDefault();
    const form = e.target as HTMLFormElement;
    
    if (form.getAttribute('data-form') === 'add-task') {
      this.handleAddTask(form);
    }
  }

  private showAddTaskModal(): void {
    this.renderer.showAddTaskModal();
  }

  private hideAddTaskModal(): void {
    this.renderer.hideAddTaskModal();
  }

  private async handleAddTask(form: HTMLFormElement): Promise<void> {
    if (this.isSubmitting) {
      return; // Prevent duplicate submissions
    }

    this.isSubmitting = true;
    
    const formData = new FormData(form);
    const taskData = {
      title: formData.get('title') as string,
      description: formData.get('description') as string,
      task_type: formData.get('task_type') as string,
      assigned_to: formData.get('assigned_to') as string || undefined,
      family_id: 'fam1' // Default family
    };

    try {
      const newTask = await this.taskService.createTask(taskData);
      // Optimistically add to state
      stateManager.addTask(newTask);
      this.hideAddTaskModal();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to create task';
      this.showFormError(form, errorMessage);
    } finally {
      this.isSubmitting = false;
    }
  }

  private showFormError(form: HTMLFormElement, errors: any): void {
    ComponentUtils.showFormErrors(form, errors);
  }

  private async handleTaskReorder(evt: any): Promise<void> {
    const taskElement = evt.item;
    const taskContainer = taskElement.querySelector('[data-task-id]');
    const taskId = taskContainer?.getAttribute('data-task-id');
    
    if (!taskId) return;

    // Check if task was moved between different columns (assignment change)
    if (evt.from !== evt.to) {
      const sourceColumn = evt.from.closest('[data-user-tasks]');
      const targetColumn = evt.to.closest('[data-user-tasks]');
      
      if (!targetColumn) {
        return;
      }

      const sourceUserKey = sourceColumn?.getAttribute('data-user-tasks');
      const targetUserKey = targetColumn.getAttribute('data-user-tasks');

      // Determine new assignment based on target column
      let newAssignedTo: string | null = null;
      if (targetUserKey === 'unassigned') {
        newAssignedTo = null;
      } else if (targetUserKey?.startsWith('user_')) {
        const userId = targetUserKey.substring(5);
        if (userId && userId.trim() !== '') {
          newAssignedTo = userId;
        } else {
          this.renderer.renderError(`Invalid user key format: ${targetUserKey}`);
          return;
        }
      } else {
        this.renderer.renderError(`Unexpected user key format: ${targetUserKey}`);
        return;
      }

      // Note: Original task counts could be stored here for future revert logic

      // Update state first for optimistic update
      stateManager.updateTask(taskId, { assigned_to: newAssignedTo });

      try {
        await this.taskService.updateTask(taskId, { assigned_to: newAssignedTo });

      } catch (error) {
        // Revert the state on API failure
        const originalAssignedTo = sourceUserKey === 'unassigned' ? null : 
                                  sourceUserKey?.substring(5) || null;
        stateManager.updateTask(taskId, { assigned_to: originalAssignedTo });
        
        // Show error to user instead of throwing
        this.renderer.renderError('Failed to update task assignment. Please try again.');
      }
    }
  }

  private getTaskCount(userKey: string): number {
    return this.renderer.getTaskCount(userKey);
  }

  private setTaskCount(userKey: string, count: number): void {
    this.renderer.updateTaskCount(userKey, count);
  }

  private updateTaskCounts(sourceUserKey: string | null, targetUserKey: string): void {
    // Update task count for source user (decrease)
    if (sourceUserKey) {
      const currentCount = this.getTaskCount(sourceUserKey);
      this.setTaskCount(sourceUserKey, Math.max(0, currentCount - 1));
    }

    // Update task count for target user (increase)
    const currentCount = this.getTaskCount(targetUserKey);
    this.setTaskCount(targetUserKey, currentCount + 1);
  }

  private updateTaskCardAssignment(taskId: string, newAssignedTo: string | null): void {
    const taskCard = this.taskCards.get(taskId);
    if (taskCard) {
      taskCard.updateAssignment(newAssignedTo);
    }
  }

  private handleTaskUpdate(_task: Task): void {
    // Task was updated, we could update local state here if needed
    // For now, the TaskCard handles its own updates
  }

  private handleTaskDelete(taskId: string): void {
    // Update state
    stateManager.removeTask(taskId);
    
    const taskCard = this.taskCards.get(taskId);
    if (taskCard) {
      taskCard.destroy();
      this.taskCards.delete(taskId);
    }
  }

  public async refresh(): Promise<void> {
    await this.loadTasks();
  }

  public destroy(): void {
    this.taskCards.forEach(card => card.destroy());
    this.taskCards.clear();
    this.dragDropManager.destroy();
    
    // Unsubscribe from state
    if (this.unsubscribeFromState) {
      this.unsubscribeFromState();
    }
    
    // Remove event listeners
    if (this.boundHandleClick) {
      this.container.removeEventListener('click', this.boundHandleClick);
    }
    if (this.boundHandleSubmit) {
      this.container.removeEventListener('submit', this.boundHandleSubmit);
    }
  }
}