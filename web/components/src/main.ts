import { TaskManager } from './task-manager';
import { FamilySelector } from './family-selector';
import { ComponentConfig } from './types';

// Global namespace for Fam-Stack components
declare global {
  interface Window {
    FamStack: {
      TaskManager: typeof TaskManager;
      FamilySelector: typeof FamilySelector;
      init: (config: ComponentConfig) => void;
      instances: {
        taskManagers: Map<string, TaskManager>;
        familySelectors: Map<string, FamilySelector>;
      };
    };
  }
}

// Initialize components when DOM is ready
function initializeComponents(config: ComponentConfig): void {
  // Initialize Task Managers
  const taskContainers = document.querySelectorAll('[data-component="task-manager"]');
  taskContainers.forEach((container, index) => {
    const instanceId = container.getAttribute('data-instance-id') ?? `task-manager-${index}`;
    const manager = new TaskManager(container as HTMLElement, config);
    window.FamStack.instances.taskManagers.set(instanceId, manager);
  });

  // Initialize Family Selectors
  const selectorContainers = document.querySelectorAll('[data-component="family-selector"]');
  selectorContainers.forEach((container, index) => {
    const instanceId = container.getAttribute('data-instance-id') ?? `family-selector-${index}`;
    const selector = new FamilySelector(container as HTMLElement, config);
    window.FamStack.instances.familySelectors.set(instanceId, selector);
  });
}

// Auto-initialize when DOM is ready
function autoInit(): void {
  const configElement = document.querySelector('script[data-famstack-config]');
  if (!configElement) {
    return;
  }

  const config = JSON.parse(configElement.textContent ?? '{}') as ComponentConfig;
  window.FamStack.init(config);
}

// Set up global FamStack object
window.FamStack = {
  TaskManager,
  FamilySelector,
  init: (config: ComponentConfig) => {
    initializeComponents(config);
  },
  instances: {
    taskManagers: new Map(),
    familySelectors: new Map(),
  },
};

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
        window.FamStack.instances.taskManagers.set(instanceId, manager);
      });

      const selectorContainers = target.querySelectorAll('[data-component="family-selector"]');
      selectorContainers.forEach((container, index) => {
        const instanceId =
          container.getAttribute('data-instance-id') ?? `family-selector-${Date.now()}-${index}`;
        const selector = new FamilySelector(container as HTMLElement, config);
        window.FamStack.instances.familySelectors.set(instanceId, selector);
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
      const manager = window.FamStack.instances.taskManagers.get(instanceId);
      if (manager) {
        manager.destroy();
        window.FamStack.instances.taskManagers.delete(instanceId);
      }
    }
  });

  // Family selectors don't need explicit cleanup, but we should remove them from the map
  const selectorContainers = target.querySelectorAll('[data-component="family-selector"]');
  selectorContainers.forEach(container => {
    const instanceId = container.getAttribute('data-instance-id');
    if (instanceId) {
      window.FamStack.instances.familySelectors.delete(instanceId);
    }
  });
});

export { TaskManager, FamilySelector };
