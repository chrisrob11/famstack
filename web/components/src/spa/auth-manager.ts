/**
 * Authentication manager for SPA
 * Handles JWT tokens, login/logout, and auth state
 */

import { logger } from '../common/logger.js';

export interface User {
  id: string;
  name: string;
  email: string;
  role: string;
  family_id: string;
}

export interface AuthSession {
  user_id: string;
  family_id: string;
  role: string;
  original_role: string;
  expires_at: number;
  issued_at: number;
}

export interface AuthResponse {
  user: User;
  session: AuthSession;
  token: string;
  permissions: string[];
}

export interface LoginCredentials {
  email: string;
  password: string;
}

export class AuthManager {
  private token: string | null = null;
  private user: User | null = null;
  private session: AuthSession | null = null;
  private apiBase: string;
  private refreshTimer: number | undefined;

  constructor(apiBase: string = '/api/v1') {
    this.apiBase = apiBase;
    this.loadFromStorage();
    this.setupTokenRefresh();
  }

  /**
   * Check if user is currently authenticated
   */
  isAuthenticated(): boolean {
    return !!this.token && !!this.user && !this.isTokenExpired();
  }

  /**
   * Get current user
   */
  getCurrentUser(): User | null {
    return this.user;
  }

  /**
   * Get current session
   */
  getCurrentSession(): AuthSession | null {
    return this.session;
  }

  /**
   * Get current auth token
   */
  getToken(): string | null {
    return this.token;
  }

  /**
   * Login with email and password
   */
  async login(credentials: LoginCredentials): Promise<boolean> {
    try {
      const response = await fetch('/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(credentials),
      });

      if (response.ok) {
        const serverResponse = await response.json();
        logger.authInfo('Login response received from server', serverResponse);

        // Server sends token via HTTP-only cookie, not in JSON response
        // Create auth data with a placeholder token since we can't access the cookie
        const authData: AuthResponse = {
          user: serverResponse.user,
          session: serverResponse.session,
          token: 'http-only-cookie', // Placeholder since we can't access the actual cookie
          permissions: serverResponse.permissions || [],
        };

        this.setAuthData(authData);
        logger.authInfo('Auth data set, authentication status', {
          isAuthenticated: this.isAuthenticated(),
        });
        return true;
      } else {
        const error = await response.json();
        throw new Error(error.message || 'Login failed');
      }
    } catch (error) {
      logger.error('Login failed:', error);
      throw error;
    }
  }

  /**
   * Logout user
   */
  async logout(): Promise<void> {
    try {
      // Call server logout endpoint
      await fetch('/auth/logout', {
        method: 'POST',
        headers: this.getAuthHeaders(),
      });
    } catch (error) {
      logger.error('Logout failed:', error);
    } finally {
      // Always clear local state
      this.clearAuthData();
    }
  }

  /**
   * Refresh auth token
   */
  async refresh(): Promise<boolean> {
    if (!this.token) {
      return false;
    }

    try {
      const response = await fetch('/auth/refresh', {
        method: 'POST',
        headers: this.getAuthHeaders(),
      });

      if (response.ok) {
        const authData: AuthResponse = await response.json();
        this.setAuthData(authData);
        return true;
      } else {
        this.clearAuthData();
        return false;
      }
    } catch (error) {
      logger.error('Token refresh failed:', error);
      this.clearAuthData();
      return false;
    }
  }

  /**
   * Downgrade to family/shared mode
   */
  async downgrade(): Promise<boolean> {
    try {
      const response = await fetch('/auth/downgrade', {
        method: 'POST',
        headers: this.getAuthHeaders(),
      });

      if (response.ok) {
        const data = await response.json();
        if (data.session) {
          this.session = data.session;
          this.saveToStorage();
        }
        return true;
      }
      return false;
    } catch (error) {
      logger.error('Auth downgrade failed:', error);
      return false;
    }
  }

  /**
   * Upgrade to personal mode
   */
  async upgrade(password: string): Promise<boolean> {
    try {
      const response = await fetch('/auth/upgrade', {
        method: 'POST',
        headers: {
          ...this.getAuthHeaders(),
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ password }),
      });

      if (response.ok) {
        const data = await response.json();
        if (data.session) {
          this.session = data.session;
          this.saveToStorage();
        }
        return true;
      }
      return false;
    } catch (error) {
      logger.error('Auth upgrade failed:', error);
      return false;
    }
  }

  /**
   * Get headers for authenticated requests
   */
  getAuthHeaders(): Record<string, string> {
    const headers: Record<string, string> = {};

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    return headers;
  }

  /**
   * Check if token is expired
   */
  private isTokenExpired(): boolean {
    if (!this.session) {
      return true;
    }

    const now = Date.now() / 1000; // Convert to seconds
    return now >= this.session.expires_at;
  }

  /**
   * Set authentication data
   */
  private setAuthData(authData: AuthResponse): void {
    this.token = authData.token;
    this.user = authData.user;
    this.session = authData.session;

    this.saveToStorage();
    this.setupTokenRefresh();
  }

  /**
   * Clear authentication data
   */
  private clearAuthData(): void {
    this.token = null;
    this.user = null;
    this.session = null;

    localStorage.removeItem('famstack_auth');
    sessionStorage.removeItem('famstack_auth');

    if (this.refreshTimer) {
      clearTimeout(this.refreshTimer);
      this.refreshTimer = undefined;
    }
  }

  /**
   * Save auth data to storage
   */
  private saveToStorage(): void {
    const authData = {
      token: this.token,
      user: this.user,
      session: this.session,
    };

    // Use localStorage for persistent login
    localStorage.setItem('famstack_auth', JSON.stringify(authData));
  }

  /**
   * Load auth data from storage
   */
  private loadFromStorage(): void {
    try {
      const stored = localStorage.getItem('famstack_auth');
      if (stored) {
        const authData = JSON.parse(stored);

        // Validate stored data
        if (authData.token && authData.user && authData.session) {
          this.token = authData.token;
          this.user = authData.user;
          this.session = authData.session;

          // Check if token is still valid
          if (this.isTokenExpired()) {
            this.clearAuthData();
          }
        }
      }
    } catch (error) {
      logger.error('Failed to load auth from storage:', error);
      this.clearAuthData();
    }
  }

  /**
   * Setup automatic token refresh
   */
  private setupTokenRefresh(): void {
    if (this.refreshTimer) {
      clearTimeout(this.refreshTimer);
    }

    if (this.session && this.token) {
      const now = Date.now() / 1000;
      const expiresAt = this.session.expires_at;
      const refreshAt = expiresAt - 15 * 60; // Refresh 15 minutes before expiration

      if (refreshAt > now) {
        const timeoutMs = (refreshAt - now) * 1000;
        this.refreshTimer = window.setTimeout(() => {
          this.refresh();
        }, timeoutMs);
      }
    }
  }

  /**
   * Initialize auth manager
   */
  async init(): Promise<void> {
    // Check if we have a valid session by calling /auth/me
    try {
      const response = await fetch('/auth/me', {
        credentials: 'include', // Use HTTP-only cookies instead of headers
      });

      if (response.ok) {
        const data = await response.json();
        // Update auth state from server response
        const authData: AuthResponse = {
          user: data.user,
          session: data.session,
          token: 'http-only-cookie', // Placeholder since we can't access the actual cookie
          permissions: data.permissions || [],
        };
        this.setAuthData(authData);
      } else {
        // No valid session, clear any stored data
        this.clearAuthData();
      }
    } catch (error) {
      logger.error('Auth verification failed:', error);
      this.clearAuthData();
    }
  }
}
