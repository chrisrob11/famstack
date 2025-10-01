/**
 * FamilyContext - Centralized family session management
 *
 * Provides cached access to current family information
 * Reduces redundant API calls across components
 */

import { logger } from '../common/logger.js';

export interface Family {
  id: string;
  name: string;
  created_at: string;
  updated_at: string;
  created_by: string;
}

export interface FamilySession {
  family_id: string;
  family?: Family;
  user_id: string;
  role: string;
}

class FamilyContextService {
  private familyId: string | null = null;
  private session: FamilySession | null = null;
  private family: Family | null = null;
  private loading = false;
  private loadPromise: Promise<void> | null = null;
  private listeners: Set<() => void> = new Set();

  /**
   * Get current family ID (cached)
   */
  async getFamilyId(): Promise<string | null> {
    if (this.familyId) {
      return this.familyId;
    }

    await this.loadSession();
    return this.familyId;
  }

  /**
   * Get current family session (cached)
   */
  async getSession(): Promise<FamilySession | null> {
    if (this.session) {
      return this.session;
    }

    await this.loadSession();
    return this.session;
  }

  /**
   * Get current family details (cached)
   */
  async getFamily(): Promise<Family | null> {
    if (this.family) {
      return this.family;
    }

    await this.loadSession();

    // If we have family info in session, use it
    if (this.session?.family) {
      this.family = this.session.family;
      return this.family;
    }

    // Otherwise fetch family details
    if (this.familyId) {
      await this.loadFamily(this.familyId);
    }

    return this.family;
  }

  /**
   * Load session from API
   */
  private async loadSession(): Promise<void> {
    // If already loading, wait for that promise
    if (this.loadPromise) {
      return this.loadPromise;
    }

    // If already loaded, return
    if (this.session) {
      return;
    }

    this.loading = true;
    this.loadPromise = this.fetchSession();

    try {
      await this.loadPromise;
    } finally {
      this.loading = false;
      this.loadPromise = null;
    }
  }

  private async fetchSession(): Promise<void> {
    try {
      const response = await fetch('/auth/me');
      if (!response.ok) {
        throw new Error('Failed to fetch session');
      }

      const data = await response.json();
      this.session = data.session;
      this.familyId = data.session?.family_id || data.user?.family_id || null;

      // If family info is included in session
      if (data.family) {
        this.family = data.family;
      }

      this.notifyListeners();
    } catch (error) {
      logger.error('Failed to load family session:', error);
      this.session = null;
      this.familyId = null;
    }
  }

  /**
   * Load full family details
   */
  private async loadFamily(familyId: string): Promise<void> {
    try {
      const response = await fetch(`/api/v1/families/${familyId}`);
      if (!response.ok) {
        throw new Error('Failed to fetch family');
      }

      this.family = await response.json();
      this.notifyListeners();
    } catch (error) {
      logger.error('Failed to load family:', error);
      this.family = null;
    }
  }

  /**
   * Refresh/invalidate cache
   */
  async refresh(): Promise<void> {
    this.session = null;
    this.familyId = null;
    this.family = null;
    this.loadPromise = null;
    await this.loadSession();
  }

  /**
   * Update family in cache
   */
  updateFamily(family: Family): void {
    this.family = family;
    this.notifyListeners();
  }

  /**
   * Clear cache
   */
  clear(): void {
    this.session = null;
    this.familyId = null;
    this.family = null;
    this.loadPromise = null;
    this.notifyListeners();
  }

  /**
   * Subscribe to family context changes
   */
  subscribe(listener: () => void): () => void {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  }

  private notifyListeners(): void {
    this.listeners.forEach(listener => listener());
  }

  /**
   * Check if data is loaded
   */
  isLoaded(): boolean {
    return this.session !== null;
  }

  /**
   * Check if currently loading
   */
  isLoading(): boolean {
    return this.loading;
  }
}

// Export singleton instance
export const familyContext = new FamilyContextService();
