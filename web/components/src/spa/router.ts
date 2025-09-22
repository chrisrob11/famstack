/**
 * Client-side router for SPA navigation
 */

import { logger } from '../common/logger.js';

export interface Route {
  path: string;
  component: string;
  requiresAuth: boolean;
  title: string;
}

export interface RouterConfig {
  routes: Route[];
  defaultRoute: string;
  loginRoute: string;
}

export class SPARouter {
  private routes: Map<string, Route> = new Map();
  private currentRoute?: Route;
  private config: RouterConfig;
  private authManager: any; // Will be injected

  constructor(config: RouterConfig) {
    this.config = config;
    this.setupRoutes();
    this.setupEventListeners();
  }

  private setupRoutes(): void {
    this.config.routes.forEach(route => {
      this.routes.set(route.path, route);
    });
  }

  private setupEventListeners(): void {
    // Handle browser back/forward buttons
    window.addEventListener('popstate', () => {
      this.navigate(window.location.pathname, false);
    });

    // Handle internal navigation links
    document.addEventListener('click', e => {
      const target = e.target as HTMLElement;
      const link = target.closest('a[data-route]');

      if (link) {
        e.preventDefault();
        const path = link.getAttribute('href') || link.getAttribute('data-route');
        if (path) {
          this.navigate(path);
        }
      }
    });
  }

  setAuthManager(authManager: any): void {
    this.authManager = authManager;
  }

  async navigate(path: string, pushState: boolean = true): Promise<void> {
    logger.routeInfo(`Navigate called with path: ${path}`);

    const route = this.findRoute(path);

    if (!route) {
      logger.warn(`Route not found: ${path}`);
      this.navigate(this.config.defaultRoute);
      return;
    }

    logger.routeInfo(`Route found: ${route.title}`, { requiresAuth: route.requiresAuth });

    // Check authentication
    const isAuthenticated = this.authManager?.isAuthenticated();
    logger.authInfo(`Authentication check`, { isAuthenticated });

    if (route.requiresAuth && !isAuthenticated) {
      logger.authInfo('Auth required but not authenticated, redirecting to login');
      this.navigate(this.config.loginRoute);
      return;
    }

    // If trying to access login while authenticated, redirect to default
    if (path === this.config.loginRoute && this.authManager?.isAuthenticated()) {
      this.navigate(this.config.defaultRoute);
      return;
    }

    // Update browser history
    if (pushState) {
      window.history.pushState({ path }, route.title, path);
    }

    // Update page title
    document.title = `${route.title} - FamStack`;

    // Load the component
    await this.loadComponent(route);
    this.currentRoute = route;

    // Update navigation active states
    this.updateNavigationState(path);
  }

  private findRoute(path: string): Route | undefined {
    // Exact match first
    if (this.routes.has(path)) {
      return this.routes.get(path);
    }

    // Pattern matching for dynamic routes (future enhancement)
    for (const [routePath, route] of this.routes) {
      if (this.matchRoute(routePath, path)) {
        return route;
      }
    }

    return undefined;
  }

  private matchRoute(routePath: string, actualPath: string): boolean {
    // Simple pattern matching - can be enhanced for dynamic segments
    const routeParts = routePath.split('/');
    const pathParts = actualPath.split('/');

    if (routeParts.length !== pathParts.length) {
      return false;
    }

    return routeParts.every((part, index) => {
      return part === pathParts[index] || part.startsWith(':');
    });
  }

  private async loadComponent(route: Route): Promise<void> {
    const appContainer = document.getElementById('app');
    if (!appContainer) {
      throw new Error('App container not found');
    }

    try {
      // Show loading state
      appContainer.innerHTML = `
        <div class="app-loading">
          <div class="loading-spinner"></div>
          <p>Loading ${route.title}...</p>
        </div>
      `;

      // Dynamic import of the component
      const componentModule = await import(`../pages/${route.component}.js`);
      const ComponentClass = componentModule.default || componentModule[route.component];

      if (!ComponentClass) {
        throw new Error(`Component ${route.component} not found`);
      }

      // Clear container and create component
      appContainer.innerHTML = '';

      // Create the page component
      const component = new ComponentClass(appContainer, this.getComponentConfig());
      await component.init();
    } catch (error) {
      logger.error('Failed to load component:', error);
      this.showError(`Failed to load ${route.title}`);
    }
  }

  private getComponentConfig(): any {
    const configElement = document.querySelector('script[data-famstack-config]');
    return configElement ? JSON.parse(configElement.textContent || '{}') : {};
  }

  private updateNavigationState(currentPath: string): void {
    // Update navigation active states
    document.querySelectorAll('[data-route]').forEach(link => {
      const linkPath = link.getAttribute('href') || link.getAttribute('data-route');
      if (linkPath === currentPath) {
        link.classList.add('active');
      } else {
        link.classList.remove('active');
      }
    });
  }

  private showError(message: string): void {
    const appContainer = document.getElementById('app');
    if (appContainer) {
      appContainer.innerHTML = `
        <div class="error-boundary">
          <h2>Error</h2>
          <p>${message}</p>
          <button onclick="window.location.reload()">Reload</button>
        </div>
      `;
    }
  }

  getCurrentRoute(): Route | undefined {
    return this.currentRoute;
  }

  init(): void {
    // Initialize with current path
    const currentPath = window.location.pathname;
    this.navigate(currentPath, false);
  }
}

// Default route configuration
export const defaultRoutes: Route[] = [
  { path: '/', component: 'daily-page', requiresAuth: true, title: 'Daily Tasks' },
  { path: '/tasks', component: 'daily-page', requiresAuth: true, title: 'Daily Tasks' },
  { path: '/daily', component: 'daily-page', requiresAuth: true, title: 'Daily Tasks' },
  { path: '/family', component: 'family-page', requiresAuth: true, title: 'Family' },
  { path: '/family/setup', component: 'family-page', requiresAuth: true, title: 'Family' },
  { path: '/schedules', component: 'schedules', requiresAuth: true, title: 'Schedules' },
  {
    path: '/integrations',
    component: 'integrations-page',
    requiresAuth: true,
    title: 'Integrations',
  },
  { path: '/login', component: 'login-page', requiresAuth: false, title: 'Login' },
  { path: '/calendar-dev', component: 'calendar-dev-page', requiresAuth: false, title: 'Calendar Development' },
];

export const routerConfig: RouterConfig = {
  routes: defaultRoutes,
  defaultRoute: '/tasks',
  loginRoute: '/login',
};
