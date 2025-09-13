import { ComponentConfig } from '../common/types.js';

export interface TaskSchedule {
  id: string;
  family_id: string;
  created_by: string;
  title: string;
  description: string;
  task_type: 'todo' | 'chore' | 'appointment';
  assigned_to?: string | null;
  days_of_week: string[];
  time_of_day?: string | null;
  priority: number;
  points: number;
  active: boolean;
  created_at: string;
}

export interface CreateScheduleRequest {
  title: string;
  description: string;
  task_type: 'todo' | 'chore' | 'appointment';
  assigned_to?: string | null;
  days_of_week: string[];
  time_of_day?: string | null;
  priority: number;
  points: number;
  family_id: string;
}

export class ScheduleService {
  private config: ComponentConfig;

  constructor(config: ComponentConfig) {
    this.config = config;
  }

  async listSchedules(familyId: string = 'fam1'): Promise<TaskSchedule[]> {
    const response = await fetch(`${this.config.apiBaseUrl}/schedules?family_id=${familyId}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': this.config.csrfToken,
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch schedules: ${response.statusText}`);
    }

    return response.json();
  }

  async createSchedule(scheduleData: CreateScheduleRequest): Promise<TaskSchedule> {
    const response = await fetch(`${this.config.apiBaseUrl}/schedules`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': this.config.csrfToken,
      },
      body: JSON.stringify(scheduleData),
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
