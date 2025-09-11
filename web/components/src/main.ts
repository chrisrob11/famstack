import { PageManager } from './pages/page-factory.js';
import { ComponentConfig } from './common/types.js';

/**
 * Page-centric architecture - main entry point
 * Pages are responsible for managing their own components
 */

class ComponentRegistry {
  private pageManagers = new Map<string, PageManager>();

  addPageManager(id: string, manager: PageManager) {
    this.pageManagers.set(id, manager);
  }

  removePageManager(id: string) {
    const manager = this.pageManagers.get(id);
    if (manager) {
      manager.destroy();
      this.pageManagers.delete(id);
    }
  }
}

const registry = new ComponentRegistry();

// Initialize pages only - no global component scanning
function initializeComponents(config: ComponentConfig): void {
  // Initialize Page Managers only - they manage their own components
  const pageContainers = document.querySelectorAll('[data-component="page"]');
  
  pageContainers.forEach((container, index) => {
    const instanceId = container.getAttribute('data-instance-id') ?? `page-${index}`;
    const pageType = container.getAttribute('data-page-type') ?? 'tasks';
    
    const pageManager = new PageManager(container as HTMLElement, config);
    pageManager.navigateToPage(pageType);
    registry.addPageManager(instanceId, pageManager);
  });
}

// Auto-initialize when DOM is ready
function autoInit(): void {
  const configElement = document.querySelector('script[data-famstack-config]');
  if (!configElement) {
    return;
  }
  
  const config: ComponentConfig = JSON.parse(configElement.textContent ?? '{}');
  initializeComponents(config);
}

// Initialize on DOM ready
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', autoInit);
} else {
  autoInit();
}

// Handle HTMX page swaps (clean page component management)
document.addEventListener('htmx:afterSwap', (event: any) => {
  const target = event.detail.target;
  
  if (target.querySelector('[data-component="page"]')) {
    const configElement = document.querySelector('script[data-famstack-config]');
    if (configElement) {
      const config: ComponentConfig = JSON.parse(configElement.textContent ?? '{}');
      
      const pageContainers = target.querySelectorAll('[data-component="page"]');
      pageContainers.forEach((container: Element, index: number) => {
        const instanceId = container.getAttribute('data-instance-id') ?? `page-${Date.now()}-${index}`;
        const pageType = container.getAttribute('data-page-type') ?? 'tasks';
        
        const pageManager = new PageManager(container as HTMLElement, config);
        pageManager.navigateToPage(pageType);
        registry.addPageManager(instanceId, pageManager);
      });
    }
  }
});

document.addEventListener('htmx:beforeSwap', (event: any) => {
  const target = event.detail.target;
  const pageContainers = target.querySelectorAll('[data-component="page"]');
  
  pageContainers.forEach((container: Element) => {
    const instanceId = container.getAttribute('data-instance-id');
    if (instanceId) {
      registry.removePageManager(instanceId);
    }
  });
});