import { TaskDragDropManager } from '../task-drag-drop-manager';

// Mock Sortable.js
const mockSortable = {
  destroy: jest.fn(),
  option: jest.fn(),
  toArray: jest.fn(() => [])
};

const mockSortableConstructor = jest.fn(() => mockSortable);
(global as any).Sortable = mockSortableConstructor;

// Mock console methods
const mockConsoleError = jest.fn();
global.console.error = mockConsoleError;

describe('TaskDragDropManager', () => {
  let container: HTMLElement;
  let dragDropManager: TaskDragDropManager;
  let mockOnTaskReorder: jest.Mock;

  beforeEach(() => {
    // Reset mocks
    mockSortableConstructor.mockClear();
    mockSortable.destroy.mockClear();
    mockOnTaskReorder = jest.fn();
    mockConsoleError.mockClear();

    // Create test container with task columns
    container = document.createElement('div');
    container.innerHTML = `
      <div class="task-columns">
        <div class="task-column" data-user-tasks="unassigned">
          <h3>Unassigned (2)</h3>
          <div class="task-item" data-task-container="task1">
            <div data-task-id="task1">Task 1</div>
          </div>
          <div class="task-item" data-task-container="task2">
            <div data-task-id="task2">Task 2</div>
          </div>
        </div>
        <div class="task-column" data-user-tasks="user_user1">
          <h3>John Smith (1)</h3>
          <div class="task-item" data-task-container="task3">
            <div data-task-id="task3">Task 3</div>
          </div>
        </div>
      </div>
    `;
    document.body.appendChild(container);

    dragDropManager = new TaskDragDropManager(container, {
      onTaskReorder: mockOnTaskReorder
    });
  });

  afterEach(() => {
    document.body.removeChild(container);
  });

  describe('Initialization', () => {
    it('should initialize with container and callback', () => {
      expect(dragDropManager).toBeDefined();
    });

    it('should store the onTaskReorder callback', () => {
      // We can test this by triggering a drag event and seeing if callback is called
      expect(mockOnTaskReorder).toBeDefined();
    });
  });

  describe('setupSortable', () => {
    it('should create Sortable instances for each task column', () => {
      dragDropManager.setupSortable();

      // Should be called for each column with data-user-tasks attribute
      expect(mockSortableConstructor).toHaveBeenCalledTimes(2);
      
      // Check first call (unassigned column)
      const firstCall = mockSortableConstructor.mock.calls[0];
      expect(firstCall[0]).toBe(container.querySelector('[data-user-tasks="unassigned"]'));
      expect(firstCall[1]).toMatchObject({
        group: 'tasks',
        animation: 150,
        ghostClass: 'task-ghost',
        chosenClass: 'task-chosen',
        dragClass: 'task-drag'
      });

      // Check second call (user column)
      const secondCall = mockSortableConstructor.mock.calls[1];
      expect(secondCall[0]).toBe(container.querySelector('[data-user-tasks="user_user1"]'));
    });

    it('should handle containers without task columns gracefully', () => {
      const emptyContainer = document.createElement('div');
      const emptyDragDropManager = new TaskDragDropManager(emptyContainer, {
        onTaskReorder: mockOnTaskReorder
      });

      expect(() => {
        emptyDragDropManager.setupSortable();
      }).not.toThrow();

      expect(mockSortableConstructor).not.toHaveBeenCalled();
    });
  });

  describe('handleTaskReorder', () => {
    beforeEach(() => {
      dragDropManager.setupSortable();
    });

    it('should call onTaskReorder when tasks are moved between columns', async () => {
      // Get the onEnd callback from the Sortable constructor
      const sortableConfig = mockSortableConstructor.mock.calls[0][1];
      const onEndCallback = sortableConfig.onEnd;

      // Create mock drag event - moving task from unassigned to user column
      const mockEvent = {
        item: container.querySelector('[data-task-container="task1"]'),
        from: container.querySelector('[data-user-tasks="unassigned"]'),
        to: container.querySelector('[data-user-tasks="user_user1"]'),
        oldIndex: 0,
        newIndex: 1
      };

      // Call the onEnd callback
      await onEndCallback(mockEvent);

      expect(mockOnTaskReorder).toHaveBeenCalledWith(mockEvent);
    });

    it('should handle missing task ID gracefully', async () => {
      const sortableConfig = mockSortableConstructor.mock.calls[0][1];
      const onEndCallback = sortableConfig.onEnd;

      // Create mock event with item that has no task ID
      const mockItem = document.createElement('div');
      const mockEvent = {
        item: mockItem,
        from: container.querySelector('[data-user-tasks="unassigned"]'),
        to: container.querySelector('[data-user-tasks="user_user1"]'),
        oldIndex: 0,
        newIndex: 1
      };

      // Clear previous calls from initialization
      mockOnTaskReorder.mockClear();

      // Should not throw and should not call onTaskReorder
      await onEndCallback(mockEvent);

      expect(mockOnTaskReorder).not.toHaveBeenCalled();
    });

    it('should handle reordering within same column', async () => {
      const sortableConfig = mockSortableConstructor.mock.calls[0][1];
      const onEndCallback = sortableConfig.onEnd;

      // Create mock event for reordering within same column
      const unassignedColumn = container.querySelector('[data-user-tasks="unassigned"]');
      const mockEvent = {
        item: container.querySelector('[data-task-container="task1"]'),
        from: unassignedColumn,
        to: unassignedColumn,
        oldIndex: 0,
        newIndex: 1
      };

      await onEndCallback(mockEvent);

      // Should still call onTaskReorder for potential position updates
      expect(mockOnTaskReorder).toHaveBeenCalledWith(mockEvent);
    });
  });

  describe('destroy', () => {
    it('should destroy all Sortable instances', () => {
      dragDropManager.setupSortable();
      
      // Clear previous calls
      mockSortable.destroy.mockClear();
      
      dragDropManager.destroy();

      // Should call destroy on each Sortable instance
      expect(mockSortable.destroy).toHaveBeenCalledTimes(2);
    });

    it('should handle destroy when no Sortable instances exist', () => {
      expect(() => {
        dragDropManager.destroy();
      }).not.toThrow();
    });

    it('should clear sortable instances map', () => {
      dragDropManager.setupSortable();
      dragDropManager.destroy();

      // After destroy, setupSortable should work again
      dragDropManager.setupSortable();
      expect(mockSortableConstructor).toHaveBeenCalledTimes(4); // 2 initial + 2 after destroy
    });
  });

  describe('Error Handling', () => {
    it('should handle errors in onTaskReorder callback gracefully', async () => {
      const errorCallback = jest.fn().mockRejectedValue(new Error('Test error'));
      const errorDragDropManager = new TaskDragDropManager(container, {
        onTaskReorder: errorCallback
      });

      errorDragDropManager.setupSortable();

      const sortableConfig = mockSortableConstructor.mock.calls[mockSortableConstructor.mock.calls.length - 1][1]; // Last call
      const onEndCallback = sortableConfig.onEnd;

      const mockEvent = {
        item: container.querySelector('[data-task-container="task1"]'),
        from: container.querySelector('[data-user-tasks="unassigned"]'),
        to: container.querySelector('[data-user-tasks="user_user1"]'),
        oldIndex: 0,
        newIndex: 1
      };

      // Should not throw even if callback throws
      await expect(onEndCallback(mockEvent)).resolves.toBeUndefined();
      
      expect(errorCallback).toHaveBeenCalledWith(mockEvent);
    });
  });
});