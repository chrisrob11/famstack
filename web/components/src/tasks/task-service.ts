import { ComponentConfig } from '../common/types.js';
import { Task } from './task-types.js';

export interface TaskColumn {
  member: {
    id: string;
    name: string;
    member_type: string;
  };
  tasks: Task[] | null;
}

export interface TasksResponse {
  tasks_by_member: { [key: string]: TaskColumn };
  date: string;
}

export interface CreateTaskData {
  title: string;
  description: string;
  task_type: string;
  assigned_to?: string | null;
  family_id: string;
  due_date?: Date | null;
  frequency?: string | null;
  priority?: number;
}

/**
 * TaskService handles all API communication for tasks
 * Separation of concerns: API logic separate from UI logic
 */
export class TaskService {
  constructor(private config: ComponentConfig) {}

  async getTasks(date?: Date): Promise<TasksResponse> {
    let url = `${this.config.apiBaseUrl}/tasks`;

    // Add date parameter if provided
    if (date) {
      const dateStr = date.toISOString().split('T')[0]; // YYYY-MM-DD format
      url += `?dueDate=${dateStr}`;
    }

    const response = await fetch(url);

    if (!response.ok) {
      throw new Error(`Failed to load tasks: ${response.statusText}`);
    }

    return await response.json();
  }

  async createTask(taskData: CreateTaskData): Promise<Task> {
    // Convert Date to ISO string for API
    const apiData = {
      ...taskData,
      due_date: taskData.due_date ? taskData.due_date.toISOString() : undefined,
    };

    const response = await fetch(`${this.config.apiBaseUrl}/tasks`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': this.config.csrfToken,
      },
      body: JSON.stringify(apiData),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.details || 'Failed to create task');
    }

    return await response.json();
  }

  async updateTask(
    taskId: string,
    updates: Partial<Task> & { assigned_to?: string | null }
  ): Promise<Task> {
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
