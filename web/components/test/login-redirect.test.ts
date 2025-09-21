/**
 * Minimal test to debug login redirect issue
 */

import { AuthManager } from '../src/spa/auth-manager';
import { SPARouter, routerConfig } from '../src/spa/router';

// Mock DOM elements
Object.defineProperty(window, 'location', {
  value: { pathname: '/login' },
  writable: true
});

// Mock DOM with app container
document.body.innerHTML = '<div id="app"></div>';

// Mock fetch for login
global.fetch = jest.fn();

// Mock localStorage
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
};
Object.defineProperty(window, 'localStorage', { value: localStorageMock });
Object.defineProperty(window, 'sessionStorage', { value: localStorageMock });

// Mock famstackApp global
Object.defineProperty(window, 'famstackApp', {
  value: null,
  writable: true
});

// Mock dynamic imports to prevent component loading errors
jest.mock('../src/pages/daily-page.js', () => ({
  default: class MockDailyPage {
    constructor() {}
    init() { return Promise.resolve(); }
  }
}), { virtual: true });

jest.mock('../src/pages/login-page.js', () => ({
  default: class MockLoginPage {
    constructor() {}
    init() { return Promise.resolve(); }
  }
}), { virtual: true });

describe('Login Redirect Issue', () => {
  let authManager: AuthManager;
  let router: SPARouter;

  beforeEach(() => {
    authManager = new AuthManager('/api/v1');
    router = new SPARouter(routerConfig);
    router.setAuthManager(authManager);

    // Mock the global app
    (window as any).famstackApp = {
      getRouter: () => router
    };

    jest.clearAllMocks();
  });

  it('should redirect after successful login', async () => {
    // Mock successful login response
    const mockLoginResponse = {
      ok: true,
      json: async () => ({
        user: {
          id: '1',
          email: 'test@example.com',
          name: 'Test User',
          role: 'user',
          family_id: 'fam1'
        },
        session: {
          user_id: '1',
          family_id: 'fam1',
          role: 'user',
          original_role: 'user',
          expires_at: Math.floor(Date.now() / 1000) + 3600,
          issued_at: Math.floor(Date.now() / 1000)
        },
        permissions: ['read', 'write']
      })
    };

    (fetch as jest.Mock).mockResolvedValueOnce(mockLoginResponse);

    // Mock router navigation
    const navigateSpy = jest.spyOn(router, 'navigate');

    // Simulate login
    const success = await authManager.login({
      email: 'test@example.com',
      password: 'password'
    });

    expect(success).toBe(true);
    expect(authManager.isAuthenticated()).toBe(true);

    // Simulate the login page redirect logic
    setTimeout(() => {
      const appRouter = (window as any).famstackApp?.getRouter();
      if (appRouter) {
        appRouter.navigate('/tasks');
      }
    }, 100);

    // Wait for the timeout and check if navigate was called
    await new Promise(resolve => setTimeout(resolve, 150));

    console.log('Navigate spy calls:', navigateSpy.mock.calls);
    console.log('Auth state:', authManager.isAuthenticated());
    console.log('Global app:', (window as any).famstackApp);
  });

  it('should check if router can load daily-page component', async () => {
    // Test if the component path issue exists
    try {
      // This should fail with the same error we're seeing
      const componentModule = await import('../src/pages/daily-page.js');
      console.log('Component loaded successfully:', componentModule);
    } catch (error) {
      console.log('Component load error (expected):', error.message);

      // Try with .ts extension instead
      try {
        const componentModule = await import('../src/pages/daily-page');
        console.log('Component loaded with .ts:', componentModule);
      } catch (tsError) {
        console.log('Component load error with .ts:', tsError.message);
      }
    }
  });
});