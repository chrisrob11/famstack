import { ScheduleService, TaskSchedule, CreateScheduleRequest } from '../schedule-service.js';
import { ComponentConfig } from '../../common/types.js';

// Mock fetch globally
global.fetch = jest.fn();

describe('ScheduleService', () => {
  let scheduleService: ScheduleService;
  let mockConfig: ComponentConfig;
  
  beforeEach(() => {
    mockConfig = {
      apiBaseUrl: '/api/v1',
      csrfToken: 'test-csrf-token'
    };
    scheduleService = new ScheduleService(mockConfig);
    (fetch as jest.Mock).mockClear();
  });

  describe('listSchedules', () => {
    it('should fetch schedules successfully', async () => {
      const mockSchedules: TaskSchedule[] = [
        {
          id: 'schedule1',
          family_id: 'fam1',
          created_by: 'user1',
          title: 'Daily Exercise',
          description: 'Morning workout routine',
          task_type: 'chore',
          assigned_to: 'user1',
          days_of_week: ['monday', 'wednesday', 'friday'],
          time_of_day: '07:00',
          priority: 1,
          points: 10,
          active: true,
          created_at: '2024-01-01T00:00:00Z'
        }
      ];

      (fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockSchedules
      });

      const result = await scheduleService.listSchedules();

      expect(fetch).toHaveBeenCalledWith('/api/v1/schedules?family_id=fam1', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': 'test-csrf-token'
        }
      });
      expect(result).toEqual(mockSchedules);
    });

    it('should throw error when fetch fails', async () => {
      (fetch as jest.Mock).mockResolvedValueOnce({
        ok: false,
        statusText: 'Internal Server Error'
      });

      await expect(scheduleService.listSchedules()).rejects.toThrow(
        'Failed to fetch schedules: Internal Server Error'
      );
    });
  });

  describe('createSchedule', () => {
    it('should create a new schedule', async () => {
      const scheduleData: CreateScheduleRequest = {
        title: 'Weekly Clean',
        description: 'Clean the house',
        task_type: 'chore',
        assigned_to: 'user1',
        days_of_week: ['saturday'],
        time_of_day: '10:00',
        priority: 2,
        points: 15,
        family_id: 'fam1'
      };

      const mockCreatedSchedule: TaskSchedule = {
        id: 'schedule2',
        ...scheduleData,
        created_by: 'user1',
        active: true,
        created_at: '2024-01-01T00:00:00Z'
      };

      (fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockCreatedSchedule
      });

      const result = await scheduleService.createSchedule(scheduleData);

      expect(fetch).toHaveBeenCalledWith('/api/v1/schedules', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': 'test-csrf-token'
        },
        body: JSON.stringify(scheduleData)
      });
      expect(result).toEqual(mockCreatedSchedule);
    });
  });

  describe('toggleSchedule', () => {
    it('should toggle schedule from active to inactive', async () => {
      const mockSchedule: TaskSchedule = {
        id: 'schedule1',
        family_id: 'fam1',
        created_by: 'user1',
        title: 'Daily Exercise',
        description: 'Morning workout routine',
        task_type: 'chore',
        assigned_to: 'user1',
        days_of_week: ['monday'],
        time_of_day: '07:00',
        priority: 1,
        points: 10,
        active: true,
        created_at: '2024-01-01T00:00:00Z'
      };

      (fetch as jest.Mock)
        .mockResolvedValueOnce({
          ok: true,
          json: async () => mockSchedule
        })
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ ...mockSchedule, active: false })
        });

      const result = await scheduleService.toggleSchedule('schedule1');

      expect(fetch).toHaveBeenCalledTimes(2);
      expect(result).toEqual({ active: false });
    });
  });
});