/**
 * Auth Manager Unit Tests
 */

import { AuthManager } from '../src/spa/auth-manager';

// Mock fetch
global.fetch = jest.fn();

// Mock localStorage and sessionStorage
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
};

const sessionStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
};

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock
});

Object.defineProperty(window, 'sessionStorage', {
  value: sessionStorageMock
});

describe('AuthManager', () => {
  let authManager: AuthManager;
  const mockFetch = fetch as jest.MockedFunction<typeof fetch>;

  beforeEach(() => {
    authManager = new AuthManager('/api/v1');
    jest.clearAllMocks();
  });

  describe('login', () => {
    it('should successfully login with valid credentials', async () => {
      const mockResponse = {
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

      mockFetch.mockResolvedValueOnce(mockResponse as any);

      const result = await authManager.login({
        email: 'test@example.com',
        password: 'password'
      });

      expect(result).toBe(true);
      expect(mockFetch).toHaveBeenCalledWith('/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email: 'test@example.com',
          password: 'password'
        })
      });
    });

    it('should fail login with invalid credentials', async () => {
      const mockResponse = {
        ok: false,
        json: async () => ({
          error: 'authentication_error',
          message: 'Invalid credentials'
        })
      };

      mockFetch.mockResolvedValueOnce(mockResponse as any);

      await expect(authManager.login({
        email: 'test@example.com',
        password: 'wrong-password'
      })).rejects.toThrow('Invalid credentials');
    });

    it('should handle network errors during login', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'));

      await expect(authManager.login({
        email: 'test@example.com',
        password: 'password'
      })).rejects.toThrow('Network error');
    });
  });

  describe('isAuthenticated', () => {
    it('should return false when no token is set', () => {
      expect(authManager.isAuthenticated()).toBe(false);
    });

    it('should return true when valid token is set', () => {
      // Mock localStorage to return auth data with proper structure
      localStorageMock.getItem.mockReturnValue(JSON.stringify({
        token: 'test-token',
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
          expires_at: Math.floor(Date.now() / 1000) + 3600, // 1 hour from now in seconds
          issued_at: Math.floor(Date.now() / 1000)
        }
      }));

      authManager = new AuthManager('/api/v1'); // Reinitialize to load from storage
      expect(authManager.isAuthenticated()).toBe(true);
    });

    it('should return false when token is expired', () => {
      // Mock localStorage to return expired auth data
      localStorageMock.getItem.mockReturnValue(JSON.stringify({
        token: 'test-token',
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
          expires_at: Math.floor(Date.now() / 1000) - 3600, // 1 hour ago in seconds
          issued_at: Math.floor(Date.now() / 1000) - 7200
        }
      }));

      authManager = new AuthManager('/api/v1'); // Reinitialize to load from storage
      expect(authManager.isAuthenticated()).toBe(false);
    });
  });

  describe('logout', () => {
    it('should clear auth data on logout', async () => {
      const mockResponse = {
        ok: true,
        json: async () => ({ success: true })
      };

      mockFetch.mockResolvedValueOnce(mockResponse as any);

      await authManager.logout();

      expect(localStorageMock.removeItem).toHaveBeenCalledWith('famstack_auth');
      expect(sessionStorageMock.removeItem).toHaveBeenCalledWith('famstack_auth');
    });
  });
});