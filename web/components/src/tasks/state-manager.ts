import { TasksResponse } from './task-service.js';
import { Task } from './task-card.js';

/**
 * AppState - Central application state
 */
export interface AppState {
  tasks: TasksResponse | null;
  selectedFamily: string | null;
  loading: boolean;
  error: string | null;
  families: Array<{id: string, name: string}>;
  users: Array<{id: string, name: string, email: string, role: string}>;
}

/**
 * StateManager - Centralized state management with subscription pattern
 * Provides single source of truth for application data
 */
export class StateManager {
  private state: AppState = {
    tasks: null,
    selectedFamily: 'fam1', // Default family
    loading: false,
    error: null,
    families: [],
    users: []
  };

  private subscribers: Array<(state: AppState) => void> = [];

  /**
   * Get current state (readonly)
   */
  getState(): Readonly<AppState> {
    return { ...this.state };
  }

  /**
   * Update state and notify subscribers
   */
  setState(updates: Partial<AppState>): void {
    const previousState = { ...this.state };
    this.state = { ...this.state, ...updates };
    this.notifySubscribers();
    void previousState;
  }

  /**
   * Subscribe to state changes
   * Returns unsubscribe function
   */
  subscribe(callback: (state: AppState) => void): () => void {
    this.subscribers.push(callback);
    
    try {
      callback(this.getState());
    } catch (error) {
      console.error('Error in state subscriber:', error);
    }
    
    // Return unsubscribe function
    return () => {
      const index = this.subscribers.indexOf(callback);
      if (index > -1) {
        this.subscribers.splice(index, 1);
      }
    };
  }

  /**
   * Notify all subscribers of state changes
   */
  private notifySubscribers(): void {
    const currentState = this.getState();
    this.subscribers.forEach(callback => {
      try {
        callback(currentState);
      } catch (error) {
        console.error('Error in state subscriber:', error);
      }
    });
  }

  // Convenience methods for common state updates

  /**
   * Set loading state
   */
  setLoading(loading: boolean): void {
    this.setState({ loading, error: loading ? null : this.state.error });
  }

  /**
   * Set error state
   */
  setError(error: string | null): void {
    this.setState({ error, loading: false });
  }

  /**
   * Update tasks data
   */
  setTasks(tasks: TasksResponse): void {
    this.setState({ tasks, loading: false, error: null });
  }

  /**
   * Add a new task optimistically
   */
  addTask(task: Task): void {
    if (!this.state.tasks) return;
    
    const updatedTasks = { ...this.state.tasks };
    const userKey = task.assigned_to ? `user_${task.assigned_to}` : 'unassigned';
    
    if (updatedTasks.tasks_by_user[userKey]) {
      const currentTasks = updatedTasks.tasks_by_user[userKey].tasks || [];
      updatedTasks.tasks_by_user[userKey].tasks = [...currentTasks, task];
    }
    
    this.setState({ tasks: updatedTasks });
  }

  /**
   * Update an existing task
   */
  updateTask(taskId: string, updates: Partial<Omit<Task, 'id'>>): void {
    if (!this.state.tasks) return;
    
    // If assigned_to is changing, handle task reassignment
    if (updates.assigned_to !== undefined) {
      this.reassignTask(taskId, updates);
      return;
    }
    
    // Otherwise, just update task properties in place
    const updatedTasks = { ...this.state.tasks };
    Object.values(updatedTasks.tasks_by_user).forEach(column => {
      if (column.tasks) {
        const taskIndex = column.tasks.findIndex(t => t.id === taskId);
        if (taskIndex > -1 && column.tasks[taskIndex]) {
          const existingTask = column.tasks[taskIndex];
          column.tasks[taskIndex] = { 
            ...existingTask, 
            ...updates,
            id: existingTask.id // Ensure id is always preserved
          } as Task;
        }
      }
    });
    
    this.setState({ tasks: updatedTasks });
  }

  /**
   * Reassign a task to a different user (move between buckets)
   */
  private reassignTask(taskId: string, updates: Partial<Omit<Task, 'id'>>): void {
    if (!this.state.tasks) return;
    
    const updatedTasks = { ...this.state.tasks };
    let taskToMove: Task | null = null;
    
    // Find and remove the task from its current location
    Object.values(updatedTasks.tasks_by_user).forEach(column => {
      if (column.tasks) {
        const taskIndex = column.tasks.findIndex(t => t.id === taskId);
        if (taskIndex > -1) {
          taskToMove = { ...column.tasks[taskIndex], ...updates } as Task;
          column.tasks.splice(taskIndex, 1);
        }
      }
    });
    
    if (taskToMove) {
      // Determine target user key
      const targetUserKey = updates.assigned_to ? `user_${updates.assigned_to}` : 'unassigned';
      
      // Add task to the new location
      if (updatedTasks.tasks_by_user[targetUserKey]) {
        const currentTasks = updatedTasks.tasks_by_user[targetUserKey].tasks || [];
        updatedTasks.tasks_by_user[targetUserKey].tasks = [...currentTasks, taskToMove];
      }
    }
    
    this.setState({ tasks: updatedTasks });
  }

  /**
   * Remove a task
   */
  removeTask(taskId: string): void {
    if (!this.state.tasks) return;
    
    const updatedTasks = { ...this.state.tasks };
    
    // Remove task from all columns
    Object.values(updatedTasks.tasks_by_user).forEach(column => {
      if (column.tasks) {
        column.tasks = column.tasks.filter(t => t.id !== taskId);
      }
    });
    
    this.setState({ tasks: updatedTasks });
  }

  /**
   * Set families list
   */
  setFamilies(families: Array<{id: string, name: string}>): void {
    this.setState({ families });
  }

  /**
   * Set users list
   */
  setUsers(users: Array<{id: string, name: string, email: string, role: string}>): void {
    this.setState({ users });
  }

  /**
   * Select a family
   */
  selectFamily(familyId: string): void {
    this.setState({ selectedFamily: familyId });
  }
}

// Global state manager instance
export const stateManager = new StateManager();