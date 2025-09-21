/**
 * Main SPA application entry point
 * Replaces the old template-based approach
 */

import { SPARouter, routerConfig } from './router.js';
import { AuthManager } from './auth-manager.js';
import { NavigationComponent } from './navigation.js';
import { logger } from '../common/logger.js';

export class FamStackApp {
  private router: SPARouter;
  private authManager: AuthManager;
  private navigation?: NavigationComponent;
  private config: any;

  constructor() {
    this.config = this.loadConfig();
    this.authManager = new AuthManager(this.config.apiBase);
    this.router = new SPARouter(routerConfig);

    // Connect auth manager and router
    this.router.setAuthManager(this.authManager);
  }

  /**
   * Initialize the application
   */
  async init(): Promise<void> {
    try {
      // Initialize auth manager
      await this.authManager.init();

      // Setup global error handling
      this.setupErrorHandling();

      // Setup navigation
      await this.setupNavigation();

      // Initialize router
      this.router.init();

      // Setup global event listeners
      this.setupGlobalListeners();

      logger.info('FamStack SPA initialized successfully');
    } catch (error) {
      logger.error('Failed to initialize FamStack SPA:', error);
      this.showError('Failed to initialize application');
    }
  }

  /**
   * Setup navigation component
   */
  private async setupNavigation(): Promise<void> {
    // Create navigation container if it doesn't exist
    let navContainer = document.getElementById('navigation');
    if (!navContainer) {
      navContainer = document.createElement('nav');
      navContainer.id = 'navigation';
      document.body.insertBefore(navContainer, document.getElementById('app'));
    }

    // Initialize navigation component
    this.navigation = new NavigationComponent(navContainer, {
      authManager: this.authManager,
      router: this.router,
    });

    await this.navigation.init();
  }

  /**
   * Setup global error handling
   */
  private setupErrorHandling(): void {
    // Handle unhandled promise rejections
    window.addEventListener('unhandledrejection', event => {
      logger.error('Unhandled promise rejection:', event.reason);

      // Don't show error for auth failures (they're handled by auth manager)
      if (
        event.reason?.message?.includes('401') ||
        event.reason?.message?.includes('Unauthorized')
      ) {
        return;
      }

      this.showError('An unexpected error occurred');
    });

    // Handle JavaScript errors
    window.addEventListener('error', event => {
      logger.error('JavaScript error:', event.error);
      this.showError('An unexpected error occurred');
    });

    // Handle fetch errors globally
    this.setupFetchInterceptor();
  }

  /**
   * Setup fetch interceptor for auth and error handling
   */
  private setupFetchInterceptor(): void {
    const originalFetch = window.fetch;

    window.fetch = async (input: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
      // For cookie-based auth, we don't need to add headers
      // The browser automatically includes HTTP-only cookies
      init = init || {};
      init.credentials = 'include'; // Ensure cookies are always included

      try {
        const response = await originalFetch(input, init);

        // Handle 401 Unauthorized
        if (response.status === 401) {
          logger.warn('Unauthorized request, clearing auth');
          await this.authManager.logout();
          this.router.navigate('/login');
        }

        return response;
      } catch (error) {
        logger.error('Fetch error:', error);
        throw error;
      }
    };
  }

  /**
   * Setup global event listeners
   */
  private setupGlobalListeners(): void {
    // Listen for auth state changes
    document.addEventListener('auth-state-changed', () => {
      if (this.navigation) {
        this.navigation.updateAuthState();
      }
    });

    // Listen for keyboard shortcuts
    document.addEventListener('keydown', event => {
      this.handleKeyboardShortcuts(event);
    });

    // Handle online/offline status
    window.addEventListener('online', () => {
      logger.info('Application is online');
    });

    window.addEventListener('offline', () => {
      logger.info('Application is offline');
    });
  }

  /**
   * Handle keyboard shortcuts
   */
  private handleKeyboardShortcuts(event: KeyboardEvent): void {
    // Only handle shortcuts when not typing in inputs
    if (
      event.target instanceof HTMLInputElement ||
      event.target instanceof HTMLTextAreaElement ||
      event.target instanceof HTMLSelectElement
    ) {
      return;
    }

    // Ctrl/Cmd + K for quick navigation
    if ((event.ctrlKey || event.metaKey) && event.key === 'k') {
      event.preventDefault();
      // Could open a command palette or quick navigation
      logger.debug('Quick navigation shortcut triggered');
    }

    // Other shortcuts can be added here
    switch (event.key) {
      case '1':
        if (event.altKey) {
          event.preventDefault();
          this.router.navigate('/tasks');
        }
        break;
      case '2':
        if (event.altKey) {
          event.preventDefault();
          this.router.navigate('/family');
        }
        break;
      case '3':
        if (event.altKey) {
          event.preventDefault();
          this.router.navigate('/schedules');
        }
        break;
      case '4':
        if (event.altKey) {
          event.preventDefault();
          this.router.navigate('/integrations');
        }
        break;
    }
  }

  /**
   * Load application configuration
   */
  private loadConfig(): any {
    const configElement = document.querySelector('script[data-famstack-config]');
    if (configElement) {
      try {
        return JSON.parse(configElement.textContent || '{}');
      } catch (error) {
        logger.error('Failed to parse config:', error);
      }
    }

    // Default config
    return {
      apiBase: '/api/v1',
      authEnabled: true,
      features: {
        tasks: true,
        calendar: true,
        family: true,
        schedules: true,
        integrations: true,
      },
    };
  }

  /**
   * Show error message
   */
  private showError(message: string): void {
    const errorBoundary = document.getElementById('error-boundary');
    const errorMessage = document.getElementById('error-message');

    if (errorBoundary && errorMessage) {
      errorMessage.textContent = message;
      errorBoundary.style.display = 'block';

      // Hide app container
      const app = document.getElementById('app');
      if (app) {
        app.style.display = 'none';
      }
    }
  }

  /**
   * Get auth manager instance
   */
  getAuthManager(): AuthManager {
    return this.authManager;
  }

  /**
   * Get router instance
   */
  getRouter(): SPARouter {
    return this.router;
  }
}

// Initialize the application
const app = new FamStackApp();

// Make app available globally for debugging
(window as any).famstackApp = app;

// Initialize when DOM is ready
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', () => app.init());
} else {
  app.init();
}

// Export for module usage
export default app;
