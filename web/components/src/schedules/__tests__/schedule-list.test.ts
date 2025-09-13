import { ScheduleList } from '../schedule-list.js';
import { ScheduleService, TaskSchedule } from '../schedule-service.js';
import { ComponentConfig } from '../../common/types.js';

// Mock the ScheduleService
jest.mock('../schedule-service.js');

describe('ScheduleList', () => {
  let mockConfig: ComponentConfig;
  let container: HTMLElement;
  let mockScheduleService: jest.Mocked<ScheduleService>;
  let mockSchedules: TaskSchedule[];

  beforeEach(() => {
    mockConfig = {
      apiBaseUrl: '/api/v1',
      csrfToken: 'test-csrf-token'
    };

    container = document.createElement('div');
    document.body.appendChild(container);

    mockSchedules = [
      {
        id: 'schedule1',
        family_id: 'fam1',
        created_by: 'user1',
        title: 'Daily Exercise',
        description: 'Morning workout',
        task_type: 'chore',
        assigned_to: 'user1',
        days_of_week: ['monday', 'wednesday'],
        time_of_day: '07:00',
        priority: 1,
        points: 10,
        active: true,
        created_at: '2024-01-01T00:00:00Z'
      }
    ];

    mockScheduleService = {
      listSchedules: jest.fn().mockResolvedValue(mockSchedules),
      createSchedule: jest.fn(),
      getSchedule: jest.fn(),
      updateSchedule: jest.fn(),
      deleteSchedule: jest.fn(),
      toggleSchedule: jest.fn().mockResolvedValue({ active: false })
    } as any;

    (ScheduleService as jest.Mock).mockImplementation(() => mockScheduleService);
  });

  afterEach(() => {
    document.body.removeChild(container);
  });

  it('should initialize and load schedules', async () => {
    const scheduleList = new ScheduleList(container, mockConfig);

    expect(mockScheduleService.listSchedules).toHaveBeenCalled();
    expect(container.querySelector('.schedule-grid')).toBeTruthy();
  });

  it('should show add schedule button', () => {
    const scheduleList = new ScheduleList(container, mockConfig);
    
    const addButton = container.querySelector('[data-action="add-schedule"]') as HTMLElement;
    expect(addButton).toBeTruthy();
    expect(addButton.textContent?.trim()).toBe('+ Add Schedule');
  });

  it('should handle schedule toggle', async () => {
    const scheduleList = new ScheduleList(container, mockConfig);
    
    // Wait for initialization
    await new Promise(resolve => setTimeout(resolve, 0));
    
    // Find and trigger toggle on the first schedule
    const toggleButton = container.querySelector('[data-action="toggle"]') as HTMLElement;
    if (toggleButton) {
      toggleButton.click();
      
      expect(mockScheduleService.toggleSchedule).toHaveBeenCalledWith('schedule1');
    }
  });

  it('should handle schedule deletion with confirmation', async () => {
    // Mock window.confirm
    const originalConfirm = window.confirm;
    window.confirm = jest.fn(() => true);
    mockScheduleService.deleteSchedule.mockResolvedValue();

    const scheduleList = new ScheduleList(container, mockConfig);
    
    // Wait for initialization
    await new Promise(resolve => setTimeout(resolve, 0));
    
    const deleteButton = container.querySelector('[data-action="delete"]') as HTMLElement;
    if (deleteButton) {
      deleteButton.click();
      
      expect(mockScheduleService.deleteSchedule).toHaveBeenCalledWith('schedule1');
    }

    // Restore original confirm
    window.confirm = originalConfirm;
  });

  it('should show empty state when no schedules', async () => {
    mockScheduleService.listSchedules.mockResolvedValue([]);
    
    const scheduleList = new ScheduleList(container, mockConfig);
    
    // Wait for async operations
    await new Promise(resolve => setTimeout(resolve, 100));
    
    expect(container.textContent).toContain('No schedules created yet');
  });
});