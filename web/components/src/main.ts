import { TaskManager } from './task-manager.js';
import { FamilySelector } from './family-selector.js';
import { TaskCard, Task } from './task-card.js';
import { TaskList } from './task-list.js';
import { ComponentConfig } from './types.js';

// Component registry - avoid global window pollution
class ComponentRegistry {
  private taskManagers = new Map<string, TaskManager>();
  private familySelectors = new Map<string, FamilySelector>();
  private taskLists = new Map<string, TaskList>();

  addTaskManager(id: string, manager: TaskManager) {
    this.taskManagers.set(id, manager);
  }

  addTaskList(id: string, taskList: TaskList) {
    this.taskLists.set(id, taskList);
  }

  addFamilySelector(id: string, selector: FamilySelector) {
    this.familySelectors.set(id, selector);
  }

  getTaskManager(id: string): TaskManager | undefined {
    return this.taskManagers.get(id);
  }

  getTaskList(id: string): TaskList | undefined {
    return this.taskLists.get(id);
  }

  getFamilySelector(id: string): FamilySelector | undefined {
    return this.familySelectors.get(id);
  }

  removeTaskManager(id: string) {
    const manager = this.taskManagers.get(id);
    if (manager) {
      manager.destroy();
      this.taskManagers.delete(id);
    }
  }

  removeTaskList(id: string) {
    const taskList = this.taskLists.get(id);
    if (taskList) {
      taskList.destroy();
      this.taskLists.delete(id);
    }
  }

  removeFamilySelector(id: string) {
    this.familySelectors.delete(id);
  }
}

const registry = new ComponentRegistry();

// Initialize components when DOM is ready
function initializeComponents(config: ComponentConfig): void {
  // Initialize Task Managers
  const taskContainers = document.querySelectorAll('[data-component="task-manager"]');
  taskContainers.forEach((container, index) => {
    const instanceId = container.getAttribute('data-instance-id') ?? `task-manager-${index}`;
    const manager = new TaskManager(container as HTMLElement, config);
    registry.addTaskManager(instanceId, manager);
  });

  // Initialize Task Lists
  const taskListContainers = document.querySelectorAll('[data-component="task-list"]');
  taskListContainers.forEach((container, index) => {
    const instanceId = container.getAttribute('data-instance-id') ?? `task-list-${index}`;
    const taskList = new TaskList(container as HTMLElement, config);
    registry.addTaskList(instanceId, taskList);
  });

  // Initialize Family Selectors
  const selectorContainers = document.querySelectorAll('[data-component="family-selector"]');
  selectorContainers.forEach((container, index) => {
    const instanceId = container.getAttribute('data-instance-id') ?? `family-selector-${index}`;
    const selector = new FamilySelector(container as HTMLElement, config);
    registry.addFamilySelector(instanceId, selector);
  });
}

// Auto-initialize when DOM is ready
function autoInit(): void {
  const configElement = document.querySelector('script[data-famstack-config]');
  if (!configElement) {
    return;
  }

  const config = JSON.parse(configElement.textContent ?? '{}') as ComponentConfig;
  initializeComponents(config);
}

// Auto-initialize if DOM is already ready, or wait for it
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', autoInit);
} else {
  autoInit();
}

// Re-initialize components after HTMX content swaps
document.addEventListener('htmx:afterSwap', event => {
  const target = (event as CustomEvent).detail.target as HTMLElement;

  // Only re-initialize if the swapped content contains our components
  if (target.querySelector('[data-component]')) {
    const configElement = document.querySelector('script[data-famstack-config]');
    if (configElement) {
      const config = JSON.parse(configElement.textContent ?? '{}') as ComponentConfig;

      // Initialize components within the swapped content
      const taskContainers = target.querySelectorAll('[data-component="task-manager"]');
      taskContainers.forEach((container, index) => {
        const instanceId =
          container.getAttribute('data-instance-id') ?? `task-manager-${Date.now()}-${index}`;
        const manager = new TaskManager(container as HTMLElement, config);
        registry.addTaskManager(instanceId, manager);
      });

      const taskListContainers = target.querySelectorAll('[data-component="task-list"]');
      taskListContainers.forEach((container, index) => {
        const instanceId =
          container.getAttribute('data-instance-id') ?? `task-list-${Date.now()}-${index}`;
        const taskList = new TaskList(container as HTMLElement, config);
        registry.addTaskList(instanceId, taskList);
      });

      const selectorContainers = target.querySelectorAll('[data-component="family-selector"]');
      selectorContainers.forEach((container, index) => {
        const instanceId =
          container.getAttribute('data-instance-id') ?? `family-selector-${Date.now()}-${index}`;
        const selector = new FamilySelector(container as HTMLElement, config);
        registry.addFamilySelector(instanceId, selector);
      });
    }
  }
});

// Clean up destroyed components
document.addEventListener('htmx:beforeSwap', event => {
  const target = (event as CustomEvent).detail.target as HTMLElement;

  // Clean up task managers
  const taskContainers = target.querySelectorAll('[data-component="task-manager"]');
  taskContainers.forEach(container => {
    const instanceId = container.getAttribute('data-instance-id');
    if (instanceId) {
      registry.removeTaskManager(instanceId);
    }
  });

  // Clean up task lists
  const taskListContainers = target.querySelectorAll('[data-component="task-list"]');
  taskListContainers.forEach(container => {
    const instanceId = container.getAttribute('data-instance-id');
    if (instanceId) {
      registry.removeTaskList(instanceId);
    }
  });

  // Clean up family selectors
  const selectorContainers = target.querySelectorAll('[data-component="family-selector"]');
  selectorContainers.forEach(container => {
    const instanceId = container.getAttribute('data-instance-id');
    if (instanceId) {
      registry.removeFamilySelector(instanceId);
    }
  });
});

export { TaskManager, FamilySelector, TaskCard, TaskList };
export type { Task };
