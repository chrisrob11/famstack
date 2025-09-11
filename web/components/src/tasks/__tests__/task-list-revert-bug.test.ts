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

describe('TaskList Drag and Drop Revert Bug', () => {
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
          }
        ]
      },
      'user_user1': {
        user: { id: 'user1', name: 'John Smith', role: 'parent' },
        tasks: []
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

    // Create test container
    container = document.createElement('div');
    document.body.appendChild(container);

    config = {
      apiBaseUrl: '/api/v1',
      csrfToken: 'test-token'
    };

    // Reset state manager
    stateManager.setState({
      tasks: null,
      loading: false,
      error: null
    });
  });

  afterEach(() => {
    if (taskList) {
      taskList.destroy();
    }
    document.body.removeChild(container);
  });

  describe('Optimistic Update vs Server Response', () => {
    it('should reproduce the revert bug when API succeeds but UI reverts', async () => {
      // Set up initial state
      stateManager.setTasks(mockTasksResponse);
      
      // Mock initial load response
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockTasksResponse
      });

      taskList = new TaskList(container, config);
      await new Promise(resolve => setTimeout(resolve, 100));

      // Clear initial fetch calls
      mockFetch.mockClear();

      // Mock successful API response for task update
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          id: 'task1',
          title: 'Task 1',
          assigned_to: 'user1',
          status: 'pending'
        })
      });

      // Create realistic DOM structure
      container.innerHTML = `
        <div class="task-columns">
          <div class="task-column" data-user-tasks="unassigned">
            <h3>Unassigned <span class="task-count">(1)</span></h3>
            <div class="task-item" data-task-container="task1">
              <div data-task-id="task1">Task 1</div>
            </div>
          </div>
          <div class="task-column" data-user-tasks="user_user1">
            <h3>John Smith <span class="task-count">(0)</span></h3>
          </div>
        </div>
      `;

      const sortableConfig = mockSortableConstructor.mock.calls[0][1];
      const onEndCallback = sortableConfig.onEnd;

      const fromColumn = container.querySelector('[data-user-tasks="unassigned"]') as HTMLElement;
      const toColumn = container.querySelector('[data-user-tasks="user_user1"]') as HTMLElement;
      const taskContainer = container.querySelector('[data-task-container="task1"]') as HTMLElement;

      // Add children arrays to simulate DOM behavior
      Object.defineProperty(fromColumn, 'children', {
        value: [taskContainer],
        writable: true
      });
      Object.defineProperty(toColumn, 'children', {
        value: [],
        writable: true
      });

      // Simulate the drag event
      const mockEvent = {
        item: taskContainer,
        from: fromColumn,
        to: toColumn,
        oldIndex: 0,
        newIndex: 0
      };

      // Capture state before drag
      const stateBefore = stateManager.getState();
      console.log('State before drag:', {
        unassignedCount: stateBefore.tasks?.tasks_by_user['unassigned']?.tasks?.length,
        userCount: stateBefore.tasks?.tasks_by_user['user_user1']?.tasks?.length
      });

      // Execute the drag operation
      await onEndCallback(mockEvent);

      // Capture state after drag
      const stateAfter = stateManager.getState();
      console.log('State after drag:', {
        unassignedCount: stateAfter.tasks?.tasks_by_user['unassigned']?.tasks?.length,
        userCount: stateAfter.tasks?.tasks_by_user['user_user1']?.tasks?.length
      });

      // Check if API was called correctly
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/tasks/task1',
        expect.objectContaining({
          method: 'PATCH',
          body: JSON.stringify({ assigned_to: 'user1' })
        })
      );

      // Verify the task moved in state
      const finalUnassignedTasks = stateAfter.tasks?.tasks_by_user['unassigned']?.tasks || [];
      const finalUserTasks = stateAfter.tasks?.tasks_by_user['user_user1']?.tasks || [];

      expect(finalUnassignedTasks.length).toBe(0);
      expect(finalUserTasks.length).toBe(1);
      expect(finalUserTasks[0].id).toBe('task1');
      expect(finalUserTasks[0].assigned_to).toBe('user1');
    });

    it('should test what happens when API fails', async () => {
      // Set up initial state
      stateManager.setTasks(mockTasksResponse);
      
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockTasksResponse
      });

      taskList = new TaskList(container, config);
      await new Promise(resolve => setTimeout(resolve, 100));

      mockFetch.mockClear();

      // Mock API failure
      mockFetch.mockRejectedValueOnce(new Error('Network error'));

      // Set up DOM
      container.innerHTML = `
        <div class="task-columns">
          <div class="task-column" data-user-tasks="unassigned">
            <h3>Unassigned <span class="task-count">(1)</span></h3>
            <div class="task-item" data-task-container="task1">
              <div data-task-id="task1">Task 1</div>
            </div>
          </div>
          <div class="task-column" data-user-tasks="user_user1">
            <h3>John Smith <span class="task-count">(0)</span></h3>
          </div>
        </div>
      `;

      const sortableConfig = mockSortableConstructor.mock.calls[0][1];
      const onEndCallback = sortableConfig.onEnd;

      const fromColumn = container.querySelector('[data-user-tasks="unassigned"]') as HTMLElement;
      const toColumn = container.querySelector('[data-user-tasks="user_user1"]') as HTMLElement;
      const taskContainer = container.querySelector('[data-task-container="task1"]') as HTMLElement;

      // Add children arrays for revert logic
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

      const stateBefore = stateManager.getState();
      
      // Execute drag operation that should fail
      await onEndCallback(mockEvent);

      const stateAfter = stateManager.getState();

      // When API fails, state should revert to original
      expect(stateAfter.tasks?.tasks_by_user['unassigned']?.tasks?.length).toBe(
        stateBefore.tasks?.tasks_by_user['unassigned']?.tasks?.length
      );
    });

    it('should check for DOM revert behavior', async () => {
      // This test specifically checks if DOM manipulation is causing the visual revert
      stateManager.setTasks(mockTasksResponse);
      
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockTasksResponse
      });

      taskList = new TaskList(container, config);
      await new Promise(resolve => setTimeout(resolve, 100));

      mockFetch.mockClear();

      // Mock successful API call
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ id: 'task1', assigned_to: 'user1' })
      });

      // Create DOM with spy on insertBefore (used for reverting)
      const fromColumn = document.createElement('div');
      fromColumn.setAttribute('data-user-tasks', 'unassigned');
      const insertBeforeSpy = jest.spyOn(fromColumn, 'insertBefore');

      const toColumn = document.createElement('div');
      toColumn.setAttribute('data-user-tasks', 'user_user1');

      const taskContainer = document.createElement('div');
      taskContainer.setAttribute('data-task-container', 'task1');
      const taskElement = document.createElement('div');
      taskElement.setAttribute('data-task-id', 'task1');
      taskContainer.appendChild(taskElement);

      // Mock children array
      Object.defineProperty(fromColumn, 'children', {
        value: [],
        writable: true
      });

      const sortableConfig = mockSortableConstructor.mock.calls[0][1];
      const onEndCallback = sortableConfig.onEnd;

      const mockEvent = {
        item: taskContainer,
        from: fromColumn,
        to: toColumn,
        oldIndex: 0,
        newIndex: 0
      };

      await onEndCallback(mockEvent);

      // If DOM revert happens, insertBefore would be called
      console.log('insertBefore calls:', insertBeforeSpy.mock.calls.length);
      
      // This should be 0 for successful operation, >0 if reverting
      expect(insertBeforeSpy).not.toHaveBeenCalled();
    });
  });
});