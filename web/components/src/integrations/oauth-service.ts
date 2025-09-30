/**
 * OAuthService
 *
 * Role: Specialized service for OAuth authentication operations
 * Responsibilities:
 * - Manage OAuth provider configurations (get, update, validate)
 * - Handle OAuth flow initiation and completion
 * - Manage OAuth token operations (refresh, revoke)
 * - Provide OAuth-specific error handling
 * - Handle OAuth callback processing
 * - Maintain separation from general integration operations
 */

import { OAuthConfig } from './integration-types.js';
import { ComponentConfig } from '../common/types.js';

export class OAuthService {
  private baseUrl = '/api/v1';
  private config: ComponentConfig | undefined;

  constructor(config?: ComponentConfig) {
    this.config = config;
  }

  /**
   * Get common headers including CSRF token
   */
  private getHeaders(): Record<string, string> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };
    if (this.config?.csrfToken) {
      headers['X-CSRF-Token'] = this.config.csrfToken;
    }
    return headers;
  }

  /**
   * Get OAuth configuration for a provider
   */
  async getOAuthConfig(provider: string): Promise<OAuthConfig> {
    const response = await fetch(`${this.baseUrl}/config/oauth/${provider}`, {
      method: 'GET',
      headers: this.getHeaders(),
    });

    if (!response.ok) {
      throw new Error(`Failed to get OAuth config: ${response.statusText}`);
    }

    return response.json();
  }

  /**
   * Update OAuth configuration for a provider
   */
  async updateOAuthConfig(provider: string, config: Partial<OAuthConfig>): Promise<OAuthConfig> {
    const response = await fetch(`${this.baseUrl}/config/oauth/${provider}`, {
      method: 'PUT',
      headers: this.getHeaders(),
      body: JSON.stringify(config),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.message || `Failed to update OAuth config: ${response.statusText}`);
    }

    return response.json();
  }

  /**
   * Initiate OAuth flow for an integration
   */
  async initiateOAuth(integrationId: string): Promise<{ authorization_url: string }> {
    const response = await fetch(`${this.baseUrl}/integrations/${integrationId}/oauth/initiate`, {
      method: 'POST',
      headers: this.getHeaders(),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.message || `Failed to initiate OAuth: ${response.statusText}`);
    }

    return response.json();
  }

  /**
   * Complete OAuth flow (callback handling)
   */
  async completeOAuth(provider: string, code: string, state?: string): Promise<{ success: boolean; integration_id?: string }> {
    const response = await fetch(`${this.baseUrl}/oauth/${provider}/callback`, {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify({ code, state }),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.message || `Failed to complete OAuth: ${response.statusText}`);
    }

    return response.json();
  }

  /**
   * Revoke OAuth token for an integration
   */
  async revokeOAuth(integrationId: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/integrations/${integrationId}/oauth/revoke`, {
      method: 'POST',
      headers: this.getHeaders(),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.message || `Failed to revoke OAuth: ${response.statusText}`);
    }
  }

  /**
   * Refresh OAuth token for an integration
   */
  async refreshOAuth(integrationId: string): Promise<{ success: boolean }> {
    const response = await fetch(`${this.baseUrl}/integrations/${integrationId}/oauth/refresh`, {
      method: 'POST',
      headers: this.getHeaders(),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.message || `Failed to refresh OAuth: ${response.statusText}`);
    }

    return response.json();
  }

  /**
   * Validate OAuth configuration for a provider
   */
  async validateOAuthConfig(provider: string): Promise<{ valid: boolean; errors?: string[] }> {
    const response = await fetch(`${this.baseUrl}/config/oauth/${provider}/validate`, {
      method: 'POST',
      headers: this.getHeaders(),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.message || `Failed to validate OAuth config: ${response.statusText}`);
    }

    return response.json();
  }
}

// Export singleton instance
export const oAuthService = new OAuthService();

// Helper function to create configured instance
export function createOAuthService(config?: ComponentConfig): OAuthService {
  return new OAuthService(config);
}