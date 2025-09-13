import { ScheduleCard } from '../schedule-card.js';
import { TaskSchedule } from '../schedule-service.js';
import { ComponentConfig } from '../../common/types.js';

describe('ScheduleCard', () => {
  let mockConfig: ComponentConfig;
  let mockSchedule: TaskSchedule;
  let onScheduleUpdate: jest.Mock;
  let onScheduleDelete: jest.Mock;
  let onScheduleToggle: jest.Mock;

  beforeEach(() => {
    mockConfig = {
      apiBaseUrl: '/api/v1',
      csrfToken: 'test-csrf-token'
    };

    mockSchedule = {
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
    };

    onScheduleUpdate = jest.fn();
    onScheduleDelete = jest.fn();
    onScheduleToggle = jest.fn();
  });

  it('should create schedule card with correct content', () => {
    const scheduleCard = new ScheduleCard(mockSchedule, mockConfig, {
      onScheduleUpdate,
      onScheduleDelete,
      onScheduleToggle
    });

    const element = scheduleCard.getElement();

    expect(element.getAttribute('data-schedule-id')).toBe('schedule1');
    expect(element.classList.contains('schedule-card')).toBe(true);
    expect(element.classList.contains('active')).toBe(true);
    expect(element.textContent).toContain('Daily Exercise');
    expect(element.textContent).toContain('Morning workout routine');
  });

  it('should show inactive class for inactive schedule', () => {
    const inactiveSchedule = { ...mockSchedule, active: false };
    const scheduleCard = new ScheduleCard(inactiveSchedule, mockConfig);

    const element = scheduleCard.getElement();

    expect(element.classList.contains('inactive')).toBe(true);
    expect(element.classList.contains('active')).toBe(false);
  });

  it('should handle toggle action', () => {
    const scheduleCard = new ScheduleCard(mockSchedule, mockConfig, {
      onScheduleToggle
    });

    const element = scheduleCard.getElement();
    const toggleButton = element.querySelector('[data-action="toggle"]') as HTMLElement;
    
    toggleButton.click();

    expect(onScheduleToggle).toHaveBeenCalledWith('schedule1');
  });

  it('should handle delete action with confirmation', () => {
    // Mock window.confirm
    const originalConfirm = window.confirm;
    window.confirm = jest.fn(() => true);

    const scheduleCard = new ScheduleCard(mockSchedule, mockConfig, {
      onScheduleDelete
    });

    const element = scheduleCard.getElement();
    const deleteButton = element.querySelector('[data-action="delete"]') as HTMLElement;
    
    deleteButton.click();

    expect(window.confirm).toHaveBeenCalledWith(
      'Are you sure you want to delete this schedule? This will not affect tasks already created from it.'
    );
    expect(onScheduleDelete).toHaveBeenCalledWith('schedule1');

    // Restore original confirm
    window.confirm = originalConfirm;
  });

  it('should format time correctly', () => {
    const morningSchedule = { ...mockSchedule, time_of_day: '07:30' };
    const scheduleCard = new ScheduleCard(morningSchedule, mockConfig);
    const element = scheduleCard.getElement();
    
    expect(element.textContent).toContain('7:30 AM');
  });

  it('should update schedule content', () => {
    const scheduleCard = new ScheduleCard(mockSchedule, mockConfig, {
      onScheduleUpdate
    });

    const updatedSchedule = { ...mockSchedule, title: 'Updated Exercise', active: false };
    scheduleCard.updateSchedule(updatedSchedule);

    const element = scheduleCard.getElement();
    expect(element.textContent).toContain('Updated Exercise');
    expect(element.classList.contains('inactive')).toBe(true);
    expect(onScheduleUpdate).toHaveBeenCalledWith(updatedSchedule);
  });
});