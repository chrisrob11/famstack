import { SchedulesPage } from '../schedules-page.js';
import { ScheduleList } from '../schedule-list.js';
import { ComponentConfig } from '../../common/types.js';

// Mock the ScheduleList
jest.mock('../schedule-list.js');

describe('SchedulesPage', () => {
  let mockConfig: ComponentConfig;
  let container: HTMLElement;
  let mockScheduleList: jest.Mocked<ScheduleList>;

  beforeEach(() => {
    mockConfig = {
      apiBaseUrl: '/api/v1',
      csrfToken: 'test-csrf-token'
    };

    container = document.createElement('div');
    document.body.appendChild(container);

    mockScheduleList = {
      refresh: jest.fn(),
      destroy: jest.fn()
    } as any;

    (ScheduleList as jest.Mock).mockImplementation(() => mockScheduleList);
  });

  afterEach(() => {
    document.body.removeChild(container);
  });

  it('should initialize page with correct structure', async () => {
    const schedulesPage = new SchedulesPage(container, mockConfig);
    await schedulesPage.init();

    expect(container.querySelector('.schedules-page')).toBeTruthy();
    expect(container.querySelector('.schedules-container')).toBeTruthy();
    expect(ScheduleList).toHaveBeenCalledWith(
      container.querySelector('#main-schedule-container'),
      mockConfig
    );
  });

  it('should refresh schedule list when refresh is called', async () => {
    const schedulesPage = new SchedulesPage(container, mockConfig);
    await schedulesPage.init();
    
    if (schedulesPage.refresh) {
      await schedulesPage.refresh();
      expect(mockScheduleList.refresh).toHaveBeenCalled();
    }
  });

  it('should destroy schedule list when destroyed', async () => {
    const schedulesPage = new SchedulesPage(container, mockConfig);
    await schedulesPage.init();
    
    schedulesPage.destroy();
    
    expect(mockScheduleList.destroy).toHaveBeenCalled();
    expect(container.innerHTML).toBe('');
  });
});