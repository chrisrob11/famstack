import { StateManager, AppState } from '../state-manager';
import { Task } from '../task-card';
import { TasksResponse } from '../task-service';

// Mock console methods
const mockConsoleLog = jest.fn();
global.console.log = mockConsoleLog;

describe('StateManager', () => {
  let stateManager: StateManager;

  beforeEach(() => {
    stateManager = new StateManager();
    mockConsoleLog.mockClear();
  });

  afterEach(() => {
    // Clean up any subscriptions to prevent memory leaks
    stateManager = null as any;
  });

  describe('Initial State', () => {
    it('should have correct default values', () => {
      const state = stateManager.getState();
      
      expect(state.tasks).toBeNull();
      expect(state.selectedFamily).toBe('fam1');
      expect(state.loading).toBe(false);
      expect(state.error).toBeNull();
      expect(Array.isArray(state.families)).toBe(true);
      expect(Array.isArray(state.users)).toBe(true);
      expect(state.families.length).toBe(0);
      expect(state.users.length).toBe(0);
    });

    it('should return a copy of state, not reference', () => {
      const state1 = stateManager.getState();
      const state2 = stateManager.getState();
      
      expect(state1).not.toBe(state2); // Different objects
      expect(state1).toEqual(state2); // Same content
    });
  });

  describe('setState', () => {
    it('should update state correctly', () => {
      stateManager.setState({ loading: true });
      
      const state = stateManager.getState();
      expect(state.loading).toBe(true);
      expect(state.selectedFamily).toBe('fam1'); // Should remain unchanged
    });


    it('should handle multiple property updates', () => {
      stateManager.setState({ 
        loading: true, 
        error: 'Test error',
        selectedFamily: 'fam2' 
      });
      
      const state = stateManager.getState();
      expect(state.loading).toBe(true);
      expect(state.error).toBe('Test error');
      expect(state.selectedFamily).toBe('fam2');
    });
  });

  describe('Subscriptions', () => {
    it('should call callback immediately on subscribe', () => {
      const callback = jest.fn();
      
      stateManager.subscribe(callback);
      
      expect(callback).toHaveBeenCalledWith(expect.objectContaining({
        tasks: null,
        selectedFamily: 'fam1',
        loading: false
      }));
    });

    it('should call callback on state updates', () => {
      const callback = jest.fn();
      stateManager.subscribe(callback);
      
      callback.mockClear();
      stateManager.setState({ loading: true });
      
      expect(callback).toHaveBeenCalledWith(expect.objectContaining({
        loading: true
      }));
    });

    it('should unsubscribe correctly', () => {
      const callback = jest.fn();
      const unsubscribe = stateManager.subscribe(callback);
      
      callback.mockClear();
      unsubscribe();
      stateManager.setState({ loading: true });
      
      expect(callback).not.toHaveBeenCalled();
    });

    it('should handle multiple subscribers', () => {
      const callback1 = jest.fn();
      const callback2 = jest.fn();
      
      stateManager.subscribe(callback1);
      stateManager.subscribe(callback2);
      
      callback1.mockClear();
      callback2.mockClear();
      
      stateManager.setState({ loading: true });
      
      expect(callback1).toHaveBeenCalledWith(expect.objectContaining({ loading: true }));
      expect(callback2).toHaveBeenCalledWith(expect.objectContaining({ loading: true }));
    });

    it('should handle subscriber errors gracefully', () => {
      const errorCallback = jest.fn(() => { throw new Error('Test error'); });
      const goodCallback = jest.fn();
      
      stateManager.subscribe(errorCallback);
      stateManager.subscribe(goodCallback);
      
      errorCallback.mockClear();
      goodCallback.mockClear();
      
      // Should not throw, should call good callback despite error in first
      expect(() => {
        stateManager.setState({ loading: true });
      }).not.toThrow();
      
      expect(goodCallback).toHaveBeenCalledWith(expect.objectContaining({ loading: true }));
    });
  });

  describe('Convenience Methods', () => {
    it('setLoading should update loading state', () => {
      stateManager.setLoading(true);
      
      expect(stateManager.getState().loading).toBe(true);
      expect(stateManager.getState().error).toBeNull();
    });

    it('setLoading(false) should preserve existing error', () => {
      stateManager.setState({ error: 'Existing error' });
      stateManager.setLoading(false);
      
      expect(stateManager.getState().loading).toBe(false);
      expect(stateManager.getState().error).toBe('Existing error');
    });

    it('setError should update error and clear loading', () => {
      stateManager.setState({ loading: true });
      stateManager.setError('Test error');
      
      expect(stateManager.getState().error).toBe('Test error');
      expect(stateManager.getState().loading).toBe(false);
    });

    it('setTasks should update tasks and clear loading/error', () => {
      const mockTasks: TasksResponse = {
        tasks_by_user: {
          'unassigned': {
            user: { id: 'unassigned', name: 'Unassigned', role: 'system' },
            tasks: []
          }
        },
        date: 'Monday, January 1'
      };

      stateManager.setState({ loading: true, error: 'Old error' });
      stateManager.setTasks(mockTasks);
      
      const state = stateManager.getState();
      expect(state.tasks).toEqual(mockTasks);
      expect(state.loading).toBe(false);
      expect(state.error).toBeNull();
    });
  });

  describe('Task Operations', () => {
    let mockTasksResponse: TasksResponse;
    let mockTask: Task;

    beforeEach(() => {
      mockTask = {
        id: 'test-1',
        title: 'Test Task',
        description: 'Test Description', 
        status: 'pending',
        task_type: 'todo',
        created_at: '2023-01-01T00:00:00Z',
        priority: 1
      };

      mockTasksResponse = {
        tasks_by_user: {
          'unassigned': {
            user: { id: 'unassigned', name: 'Unassigned', role: 'system' },
            tasks: []
          },
          'user_1': {
            user: { id: '1', name: 'Test User', role: 'parent' },
            tasks: []
          }
        },
        date: 'Monday, January 1'
      };

      stateManager.setTasks(mockTasksResponse);
    });

    describe('addTask', () => {
      it('should add unassigned task to unassigned column', () => {
        stateManager.addTask({ ...mockTask, assigned_to: null });
        
        const state = stateManager.getState();
        const unassignedTasks = state.tasks?.tasks_by_user['unassigned']?.tasks || [];
        
        expect(unassignedTasks.length).toBe(1);
        expect(unassignedTasks[0].id).toBe('test-1');
      });

      it('should add assigned task to correct user column', () => {
        stateManager.addTask({ ...mockTask, assigned_to: '1' });
        
        const state = stateManager.getState();
        const userTasks = state.tasks?.tasks_by_user['user_1']?.tasks || [];
        
        expect(userTasks.length).toBe(1);
        expect(userTasks[0].assigned_to).toBe('1');
      });

      it('should handle missing tasks state gracefully', () => {
        const freshStateManager = new StateManager();
        expect(() => {
          freshStateManager.addTask(mockTask);
        }).not.toThrow();
      });
    });

    describe('updateTask', () => {
      beforeEach(() => {
        stateManager.addTask(mockTask);
      });

      it('should update task properties', () => {
        stateManager.updateTask('test-1', { status: 'completed' });
        
        const state = stateManager.getState();
        const task = state.tasks?.tasks_by_user['unassigned']?.tasks?.[0];
        
        expect(task?.status).toBe('completed');
        expect(task?.title).toBe('Test Task'); // Other properties preserved
      });

      it('should handle non-existent task gracefully', () => {
        expect(() => {
          stateManager.updateTask('non-existent', { status: 'completed' });
        }).not.toThrow();
      });

      it('should preserve task id during updates', () => {
        stateManager.updateTask('test-1', { title: 'Updated Title' });
        
        const state = stateManager.getState();
        const task = state.tasks?.tasks_by_user['unassigned']?.tasks?.[0];
        
        expect(task?.id).toBe('test-1');
        expect(task?.title).toBe('Updated Title');
      });
    });

    describe('removeTask', () => {
      beforeEach(() => {
        stateManager.addTask(mockTask);
        stateManager.addTask({ ...mockTask, id: 'test-2', assigned_to: '1' });
      });

      it('should remove task from correct column', () => {
        stateManager.removeTask('test-1');
        
        const state = stateManager.getState();
        const unassignedTasks = state.tasks?.tasks_by_user['unassigned']?.tasks || [];
        const userTasks = state.tasks?.tasks_by_user['user_1']?.tasks || [];
        
        expect(unassignedTasks.length).toBe(0);
        expect(userTasks.length).toBe(1); // Other task should remain
      });

      it('should handle non-existent task gracefully', () => {
        expect(() => {
          stateManager.removeTask('non-existent');
        }).not.toThrow();
      });
    });
  });

  describe('Other Operations', () => {
    it('setFamilies should update families list', () => {
      const families = [{ id: 'fam1', name: 'Test Family' }];
      stateManager.setFamilies(families);
      
      expect(stateManager.getState().families).toEqual(families);
    });

    it('setUsers should update users list', () => {
      const users = [{ id: 'user1', name: 'Test User', email: 'test@example.com', role: 'parent' }];
      stateManager.setUsers(users);
      
      expect(stateManager.getState().users).toEqual(users);
    });

    it('selectFamily should update selected family', () => {
      stateManager.selectFamily('fam2');
      
      expect(stateManager.getState().selectedFamily).toBe('fam2');
    });
  });
});