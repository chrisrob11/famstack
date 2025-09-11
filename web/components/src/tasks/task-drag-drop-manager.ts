// Use global Sortable from CDN
declare const Sortable: any;

/**
 * TaskDragDropManager - Single responsibility: Handle drag and drop functionality
 * Separated from other concerns like rendering and API calls
 */
export class TaskDragDropManager {
  private sortableInstances: Map<string, any> = new Map();
  private container: HTMLElement;
  private onTaskReorder: (evt: any) => Promise<void> = async () => {};

  constructor(container: HTMLElement, options?: {
    onTaskReorder?: (evt: any) => Promise<void>;
  }) {
    this.container = container;
    if (options?.onTaskReorder) {
      this.onTaskReorder = options.onTaskReorder;
    }
  }

  setupSortable(): void {
    // Clear existing sortable instances
    this.sortableInstances.forEach(sortable => sortable.destroy());
    this.sortableInstances.clear();

    // Check if Sortable is available (loaded from CDN)
    if (typeof Sortable === 'undefined') {
      return;
    }

    // Create sortable for each task column
    const taskColumns = this.container.querySelectorAll('[data-user-tasks]');
    taskColumns.forEach(column => {
      const userKey = column.getAttribute('data-user-tasks');
      if (userKey) {
        const sortable = new Sortable(column as HTMLElement, {
          group: 'tasks',
          animation: 150,
          ghostClass: 'task-ghost',
          chosenClass: 'task-chosen',
          dragClass: 'task-drag',
          onEnd: (evt: any) => this.handleTaskReorder(evt)
        });
        this.sortableInstances.set(userKey, sortable);
      }
    });
  }

  private async handleTaskReorder(evt: any): Promise<void> {
    // Basic validation - ensure we have a task ID
    const taskContainer = evt.item?.querySelector('[data-task-id]');
    const taskId = taskContainer?.getAttribute('data-task-id');
    
    if (!taskId) {
      // No task ID found, don't proceed with reorder
      return;
    }
    
    try {
      await this.onTaskReorder(evt);
    } catch (error) {
      console.error('Error in task reorder callback:', error);
    }
  }

  destroy(): void {
    this.sortableInstances.forEach(sortable => sortable.destroy());
    this.sortableInstances.clear();
  }

  /**
   * Get sortable instance for a specific user key
   */
  getSortableInstance(userKey: string): any {
    return this.sortableInstances.get(userKey);
  }

  /**
   * Check if drag and drop is available
   */
  isAvailable(): boolean {
    return typeof Sortable !== 'undefined';
  }
}