import { ComponentConfig } from '../common/types.js';
import { Task } from './task-card.js';

export interface TaskColumn {
  user: {
    id: string;
    name: string;
    role: string;
  };
  tasks: Task[] | null;
}

export interface TasksResponse {
  tasks_by_user: { [key: string]: TaskColumn };
  date: string;
}

export interface CreateTaskData {
  title: string;
  description: string;
  task_type: string;
  assigned_to?: string | undefined;
  family_id: string;
}

/**
 * TaskService handles all API communication for tasks
 * Separation of concerns: API logic separate from UI logic
 */
export class TaskService {
  constructor(private config: ComponentConfig) {}

  async getTasks(): Promise<TasksResponse> {
    const response = await fetch(`${this.config.apiBaseUrl}/tasks`);
    
    if (!response.ok) {
      throw new Error(`Failed to load tasks: ${response.statusText}`);
    }
    
    return await response.json();
  }

  async createTask(taskData: CreateTaskData): Promise<Task> {
    const response = await fetch(`${this.config.apiBaseUrl}/tasks`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': this.config.csrfToken,
      },
      body: JSON.stringify(taskData),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.details || 'Failed to create task');
    }

    return await response.json();
  }

  async updateTask(taskId: string, updates: Partial<Task> & { assigned_to?: string | null }): Promise<Task> {
    const response = await fetch(`${this.config.apiBaseUrl}/tasks/${taskId}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': this.config.csrfToken,
      },
      body: JSON.stringify(updates),
    });

    if (!response.ok) {
      throw new Error(`Failed to update task: ${response.statusText}`);
    }

    return await response.json();
  }

  async deleteTask(taskId: string): Promise<void> {
    const response = await fetch(`${this.config.apiBaseUrl}/tasks/${taskId}`, {
      method: 'DELETE',
      headers: {
        'X-CSRF-Token': this.config.csrfToken,
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to delete task: ${response.statusText}`);
    }
  }
}