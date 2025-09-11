import { TaskList } from '../task-list';
import { stateManager } from '../state-manager';
import { TasksResponse } from '../task-service';

// Mock Sortable.js
const mockSortable = {
  destroy: jest.fn(),
  option: jest.fn(),
  toArray: jest.fn(() => [])
};

const mockSortableConstructor = jest.fn(() => mockSortable);
(global as any).Sortable = mockSortableConstructor;

// Mock fetch for TaskService
const mockFetch = jest.fn();
global.fetch = mockFetch;

// Mock console methods
const mockConsoleLog = jest.fn();
const mockConsoleError = jest.fn();
global.console.log = mockConsoleLog;
global.console.error = mockConsoleError;

describe('TaskList Drag and Drop Integration', () => {
  let container: HTMLElement;
  let taskList: TaskList;
  let config: any;

  const mockTasksResponse: TasksResponse = {
    tasks_by_user: {
      'unassigned': {
        user: { id: 'unassigned', name: 'Unassigned', role: 'system' },
        tasks: [
          {
            id: 'task1',
            title: 'Task 1',
            description: 'Description 1',
            status: 'pending',
            task_type: 'todo',
            created_at: '2023-01-01T00:00:00Z',
            priority: 1,
            assigned_to: null
          },
          {
            id: 'task2', 
            title: 'Task 2',
            description: 'Description 2',
            status: 'pending',
            task_type: 'chore',
            created_at: '2023-01-01T01:00:00Z',
            priority: 1,
            assigned_to: null
          }
        ]
      },
      'user_user1': {
        user: { id: 'user1', name: 'John Smith', role: 'parent' },
        tasks: [
          {
            id: 'task3',
            title: 'Task 3', 
            description: 'Description 3',
            status: 'pending',
            task_type: 'todo',
            created_at: '2023-01-01T02:00:00Z',
            priority: 1,
            assigned_to: 'user1'
          }
        ]
      }
    },
    date: 'Monday, January 1'
  };

  beforeEach(() => {
    // Reset mocks
    mockSortableConstructor.mockClear();
    mockSortable.destroy.mockClear();
    mockFetch.mockClear();
    mockConsoleLog.mockClear();
    mockConsoleError.mockClear();

    // Mock successful API responses
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => mockTasksResponse
    });

    // Create test container
    container = document.createElement('div');
    document.body.appendChild(container);

    config = {
      apiBaseUrl: '/api/v1',
      csrfToken: 'test-token'
    };
  });

  afterEach(() => {
    if (taskList) {
      taskList.destroy();
    }
    document.body.removeChild(container);
    
    // Reset state manager
    stateManager.setState({
      tasks: null,
      loading: false,
      error: null
    });
  });

  describe('Drag and Drop Setup', () => {
    it('should initialize drag and drop when tasks are loaded', async () => {
      taskList = new TaskList(container, config);
      
      // Wait for tasks to load
      await new Promise(resolve => setTimeout(resolve, 100));
      
      // Should have called setupSortable after rendering tasks
      expect(mockSortableConstructor).toHaveBeenCalled();
    });

    it('should create sortable instances for each task column', async () => {
      taskList = new TaskList(container, config);
      
      // Wait for tasks to load and render
      await new Promise(resolve => setTimeout(resolve, 100));

      // Should create Sortable for each column that has tasks
      expect(mockSortableConstructor).toHaveBeenCalledTimes(2);
      
      // Verify sortable configuration
      const firstCall = mockSortableConstructor.mock.calls[0];
      expect(firstCall[1]).toMatchObject({
        group: 'tasks',
        animation: 150
      });
    });
  });

  describe('Task Reordering Integration', () => {
    beforeEach(async () => {
      taskList = new TaskList(container, config);
      
      // Wait for initial load
      await new Promise(resolve => setTimeout(resolve, 100));
      
      // Clear fetch calls from initial load
      mockFetch.mockClear();
    });

    it('should handle moving task from unassigned to user column', async () => {
      // Mock successful task update API call
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          id: 'task1',
          title: 'Task 1',
          assigned_to: 'user1',
          status: 'pending'
        })
      });

      // Get the onEnd callback from Sortable
      const sortableConfig = mockSortableConstructor.mock.calls[0][1];
      const onEndCallback = sortableConfig.onEnd;

      // Create DOM elements that would exist after rendering
      const taskContainer = document.createElement('div');
      taskContainer.setAttribute('data-task-container', 'task1');
      
      const taskElement = document.createElement('div');
      taskElement.setAttribute('data-task-id', 'task1');
      taskContainer.appendChild(taskElement);

      const fromColumn = document.createElement('div');
      fromColumn.setAttribute('data-user-tasks', 'unassigned');
      
      const toColumn = document.createElement('div');
      toColumn.setAttribute('data-user-tasks', 'user_user1');

      // Simulate drag event
      const mockEvent = {
        item: taskContainer,
        from: fromColumn,
        to: toColumn,
        oldIndex: 0,
        newIndex: 0
      };

      // Trigger the drag handler
      await onEndCallback(mockEvent);

      // Should call API to update task assignment
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/tasks/task1',
        expect.objectContaining({
          method: 'PATCH',
          headers: expect.objectContaining({
            'Content-Type': 'application/json'
          }),
          body: JSON.stringify({ assigned_to: 'user1' })
        })
      );
    });

    it('should handle moving task to unassigned column', async () => {
      // Mock successful task update API call
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          id: 'task3',
          title: 'Task 3',
          assigned_to: null,
          status: 'pending'
        })
      });

      const sortableConfig = mockSortableConstructor.mock.calls[0][1];
      const onEndCallback = sortableConfig.onEnd;

      // Create DOM elements for task moving to unassigned
      const taskContainer = document.createElement('div');
      taskContainer.setAttribute('data-task-container', 'task3');
      
      const taskElement = document.createElement('div');
      taskElement.setAttribute('data-task-id', 'task3');
      taskContainer.appendChild(taskElement);

      const fromColumn = document.createElement('div');
      fromColumn.setAttribute('data-user-tasks', 'user_user1');
      
      const toColumn = document.createElement('div');
      toColumn.setAttribute('data-user-tasks', 'unassigned');

      const mockEvent = {
        item: taskContainer,
        from: fromColumn,
        to: toColumn,
        oldIndex: 0,
        newIndex: 0
      };

      await onEndCallback(mockEvent);

      // Should call API to unassign task
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/tasks/task3',
        expect.objectContaining({
          method: 'PATCH',
          body: JSON.stringify({ assigned_to: null })
        })
      );
    });

    it('should handle API errors during drag and drop', async () => {
      // Mock API failure
      mockFetch.mockRejectedValueOnce(new Error('Network error'));

      const sortableConfig = mockSortableConstructor.mock.calls[0][1];
      const onEndCallback = sortableConfig.onEnd;

      // Create DOM elements
      const taskContainer = document.createElement('div');
      taskContainer.setAttribute('data-task-container', 'task1');
      
      const taskElement = document.createElement('div');
      taskElement.setAttribute('data-task-id', 'task1');
      taskContainer.appendChild(taskElement);

      const fromColumn = document.createElement('div');
      fromColumn.setAttribute('data-user-tasks', 'unassigned');
      
      const toColumn = document.createElement('div');
      toColumn.setAttribute('data-user-tasks', 'user_user1');

      // Mock children array using Object.defineProperty
      Object.defineProperty(fromColumn, 'children', {
        value: [],
        writable: true,
        configurable: true
      });

      const mockEvent = {
        item: taskContainer,
        from: fromColumn,
        to: toColumn,
        oldIndex: 0,
        newIndex: 0
      };

      // Should not throw even if API fails
      await expect(onEndCallback(mockEvent)).resolves.toBeUndefined();
      
      // Should attempt API call
      expect(mockFetch).toHaveBeenCalled();
    });

    it('should validate user key format', async () => {
      // Ensure we have a fresh setup for this test
      const sortableConfig = mockSortableConstructor.mock.calls[mockSortableConstructor.mock.calls.length - 1][1];
      const onEndCallback = sortableConfig.onEnd;

      // Create DOM elements with invalid user key
      const taskContainer = document.createElement('div');
      taskContainer.setAttribute('data-task-container', 'task1');
      
      const taskElement = document.createElement('div');
      taskElement.setAttribute('data-task-id', 'task1');
      taskContainer.appendChild(taskElement);

      const fromColumn = document.createElement('div');
      fromColumn.setAttribute('data-user-tasks', 'unassigned');
      
      const toColumn = document.createElement('div');
      toColumn.setAttribute('data-user-tasks', 'invalid_format'); // Invalid format

      const mockEvent = {
        item: taskContainer,
        from: fromColumn,
        to: toColumn,
        oldIndex: 0,
        newIndex: 0
      };

      await onEndCallback(mockEvent);

      // Should not call API for invalid user format
      expect(mockFetch).not.toHaveBeenCalled();
    });
  });

  describe('State Manager Integration', () => {
    it('should update state manager when drag and drop succeeds', async () => {
      // Set up initial state
      stateManager.setTasks(mockTasksResponse);
      
      taskList = new TaskList(container, config);
      await new Promise(resolve => setTimeout(resolve, 100));

      // Mock successful API call
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ id: 'task1', assigned_to: 'user1' })
      });

      const sortableConfig = mockSortableConstructor.mock.calls[0][1];
      const onEndCallback = sortableConfig.onEnd;

      // Create DOM simulation
      const taskContainer = document.createElement('div');
      taskContainer.setAttribute('data-task-container', 'task1');
      
      const taskElement = document.createElement('div');
      taskElement.setAttribute('data-task-id', 'task1');
      taskContainer.appendChild(taskElement);

      const fromColumn = document.createElement('div');
      fromColumn.setAttribute('data-user-tasks', 'unassigned');
      
      const toColumn = document.createElement('div');
      toColumn.setAttribute('data-user-tasks', 'user_user1');

      const mockEvent = {
        item: taskContainer,
        from: fromColumn,
        to: toColumn,
        oldIndex: 0,
        newIndex: 0
      };

      await onEndCallback(mockEvent);

      // Should have updated state
      const state = stateManager.getState();
      const unassignedTasks = state.tasks?.tasks_by_user['unassigned']?.tasks || [];
      const userTasks = state.tasks?.tasks_by_user['user_user1']?.tasks || [];

      // Task should have been moved from unassigned to user1
      expect(unassignedTasks.find(t => t.id === 'task1')).toBeUndefined();
      expect(userTasks.find(t => t.id === 'task1' && t.assigned_to === 'user1')).toBeTruthy();
    });
  });
});