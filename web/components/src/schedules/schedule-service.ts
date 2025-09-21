import { ComponentConfig } from '../common/types.js';

export interface TaskSchedule {
  id: string;
  family_id: string;
  created_by: string;
  title: string;
  description?: string | null;
  task_type: 'todo' | 'chore' | 'appointment';
  assigned_to?: string | null;
  days_of_week?: string | null; // JSON string that needs parsing
  time_of_day?: string | null;
  priority: number;
  points: number;
  active: boolean;
  created_at: string;
  last_generated_date?: string | null;
}

export interface CreateScheduleRequest {
  title: string;
  description?: string | null;
  task_type: 'todo' | 'chore' | 'appointment';
  assigned_to?: string | null;
  days_of_week: string[]; // Will be converted to JSON string for API
  time_of_day?: string | null;
  priority: number;
  family_id?: string; // Optional - backend will get from session if not provided
}

// Helper functions for days_of_week handling
export function parseDaysOfWeek(daysOfWeekJson?: string | null): string[] {
  if (!daysOfWeekJson) return [];
  try {
    return JSON.parse(daysOfWeekJson);
  } catch {
    return [];
  }
}

export function stringifyDaysOfWeek(daysOfWeek: string[]): string {
  return JSON.stringify(daysOfWeek);
}

export class ScheduleService {
  private config: ComponentConfig;

  constructor(config: ComponentConfig) {
    this.config = config;
  }

  async listSchedules(): Promise<TaskSchedule[]> {
    const response = await fetch(`${this.config.apiBaseUrl}/schedules`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': this.config.csrfToken,
      },
      credentials: 'include', // Include authentication cookies
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch schedules: ${response.statusText}`);
    }

    return response.json();
  }

  async createSchedule(scheduleData: CreateScheduleRequest): Promise<TaskSchedule> {
    // Send days_of_week as actual array, not JSON string
    const apiData = {
      ...scheduleData,
      days_of_week: scheduleData.days_of_week,
    };

    const response = await fetch(`${this.config.apiBaseUrl}/schedules`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': this.config.csrfToken,
      },
      credentials: 'include',
      body: JSON.stringify(apiData),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText || `Failed to create schedule: ${response.statusText}`);
    }

    return response.json();
  }

  async getSchedule(scheduleId: string): Promise<TaskSchedule> {
    const response = await fetch(`${this.config.apiBaseUrl}/schedules/${scheduleId}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': this.config.csrfToken,
      },
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch schedule: ${response.statusText}`);
    }

    return response.json();
  }

  async updateSchedule(
    scheduleId: string,
    updates: Partial<CreateScheduleRequest & { active?: boolean }>
  ): Promise<TaskSchedule> {
    const response = await fetch(`${this.config.apiBaseUrl}/schedules/${scheduleId}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': this.config.csrfToken,
      },
      credentials: 'include',
      body: JSON.stringify(updates),
    });

    if (!response.ok) {
      throw new Error(`Failed to update schedule: ${response.statusText}`);
    }

    return response.json();
  }

  async deleteSchedule(scheduleId: string): Promise<void> {
    const response = await fetch(`${this.config.apiBaseUrl}/schedules/${scheduleId}`, {
      method: 'DELETE',
      headers: {
        'X-CSRF-Token': this.config.csrfToken,
      },
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Failed to delete schedule: ${response.statusText}`);
    }
  }

  async toggleSchedule(scheduleId: string): Promise<{ active: boolean }> {
    // Get current schedule to determine current active state
    const currentSchedule = await this.getSchedule(scheduleId);
    const newActiveState = !currentSchedule.active;

    // Update the active field
    await this.updateSchedule(scheduleId, { active: newActiveState });

    return { active: newActiveState };
  }
}
