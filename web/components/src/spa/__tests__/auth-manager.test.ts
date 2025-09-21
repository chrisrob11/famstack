/**
 * Find auth manager issue
 */

import { AuthManager } from '../auth-manager';

// Mock localStorage
const mockStorage = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
};
Object.defineProperty(window, 'localStorage', { value: mockStorage });
Object.defineProperty(window, 'sessionStorage', { value: mockStorage });

global.fetch = jest.fn();

describe('AuthManager Issue', () => {
  it('login sets auth state correctly', async () => {
    const authManager = new AuthManager();

    (fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        access_token: 'token',
        user: { id: '1', email: 'test@test.com' }
      })
    });

    const result = await authManager.login({ email: 'test@test.com', password: 'pass' });

    console.log('Login result:', result);
    console.log('Is authenticated:', authManager.isAuthenticated());

    expect(result).toBe(true);
  });
});