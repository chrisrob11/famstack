/**
 * IntegrationsService
 *
 * Role: Core API service for integration CRUD operations
 * Responsibilities:
 * - Handle HTTP requests for integration data (GET, POST, PATCH, DELETE)
 * - Map backend data formats to frontend Integration objects
 * - Manage request headers and CSRF tokens
 * - Provide filtering and pagination support
 * - Handle API errors and response parsing
 * - Expose methods for sync and test operations
 */

import { Integration, OAuthConfig } from './integration-types.js';
import { ComponentConfig } from '../common/types.js';
import { logger } from '../common/logger.js';

// Backend API response types
interface BackendIntegrationResponse {
  id: string;
  family_id: string;
  created_by: string;
  integration_type: string;
  provider: string;
  auth_method: string;
  status: string;
  display_name: string;
  description: string;
  settings: string; // JSON string
  last_sync_at?: string;
  last_error?: string;
  created_at: string;
  updated_at: string;
}

interface CreateIntegrationRequest {
  integration_type: string;
  provider: string;
  auth_method: string;
  display_name: string;
  description?: string;
  settings?: Record<string, any>;
}

interface UpdateIntegrationRequest {
  display_name?: string;
  description?: string;
  settings?: Record<string, any>;
  status?: string;
}

export class IntegrationsService {
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
   * Map backend integration to frontend format
   */
  private mapIntegration(backendIntegration: BackendIntegrationResponse): Integration {
    let settings = {};
    try {
      if (backendIntegration.settings) {
        settings = JSON.parse(backendIntegration.settings);
      }
    } catch (e) {
      logger.warn('Failed to parse integration settings:', e);
    }

    return {
      id: backendIntegration.id,
      family_id: backendIntegration.family_id,
      created_by: backendIntegration.created_by,
      integration_type: backendIntegration.integration_type,
      provider: backendIntegration.provider,
      display_name: backendIntegration.display_name,
      description: backendIntegration.description,
      status: backendIntegration.status as Integration['status'],
      auth_method: backendIntegration.auth_method,
      settings,
      last_sync_at: backendIntegration.last_sync_at || null,
      last_error: backendIntegration.last_error || null,
      created_at: backendIntegration.created_at,
      updated_at: backendIntegration.updated_at,
    };
  }

  /**
   * Get all integrations with optional filters
   */
  async getIntegrations(
    filters: {
      type?: string;
      provider?: string;
      status?: string;
      limit?: number;
      offset?: number;
    } = {}
  ): Promise<Integration[]> {
    const params = new URLSearchParams();

    if (filters.type) params.append('type', filters.type);
    if (filters.provider) params.append('provider', filters.provider);
    if (filters.status) params.append('status', filters.status);
    if (filters.limit) params.append('limit', filters.limit.toString());
    if (filters.offset) params.append('offset', filters.offset.toString());

    const url = `${this.baseUrl}/integrations${params.toString() ? `?${params.toString()}` : ''}`;

    const response = await fetch(url, {
      method: 'GET',
      headers: this.getHeaders(),
    });
    if (!response.ok) {
      throw new Error(`Failed to fetch integrations: ${response.statusText}`);
    }

    const data = await response.json();

    // Handle both direct array and wrapped response
    const backendIntegrations = Array.isArray(data) ? data : data.integrations || [];

    return backendIntegrations.map((integration: BackendIntegrationResponse) =>
      this.mapIntegration(integration)
    );
  }

  /**
   * Get a single integration by ID
   */
  async getIntegration(id: string, includeCredentials = false): Promise<Integration> {
    const url = `${this.baseUrl}/integrations/${id}${includeCredentials ? '?include_credentials=true' : ''}`;

    const response = await fetch(url, {
      method: 'GET',
      headers: this.getHeaders(),
    });
    if (!response.ok) {
      throw new Error(`Failed to fetch integration: ${response.statusText}`);
    }

    const data = await response.json();

    // If credentials are included, the response structure is different
    if (includeCredentials && data.integration) {
      return this.mapIntegration(data.integration);
    }

    return this.mapIntegration(data);
  }

  /**
   * Create a new integration
   */
  async createIntegration(integrationData: Partial<Integration>): Promise<Integration> {
    const request: CreateIntegrationRequest = {
      integration_type: integrationData.integration_type!,
      provider: integrationData.provider!,
      auth_method: integrationData.auth_method!,
      display_name: integrationData.display_name!,
      ...(integrationData.description && { description: integrationData.description }),
      ...(integrationData.settings && { settings: integrationData.settings }),
    };

    const response = await fetch(`${this.baseUrl}/integrations`, {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.message || `Failed to create integration: ${response.statusText}`);
    }

    const backendIntegration = await response.json();
    return this.mapIntegration(backendIntegration);
  }

  /**
   * Update an existing integration
   */
  async updateIntegration(id: string, updates: Partial<Integration>): Promise<Integration> {
    const request: UpdateIntegrationRequest = {};

    if (updates.display_name !== undefined) request.display_name = updates.display_name;
    if (updates.description !== undefined) request.description = updates.description;
    if (updates.settings !== undefined) request.settings = updates.settings;
    if (updates.status !== undefined) request.status = updates.status;

    const response = await fetch(`${this.baseUrl}/integrations/${id}`, {
      method: 'PATCH',
      headers: this.getHeaders(),
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.message || `Failed to update integration: ${response.statusText}`);
    }

    const backendIntegration = await response.json();
    return this.mapIntegration(backendIntegration);
  }

  /**
   * Delete an integration
   */
  async deleteIntegration(id: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/integrations/${id}`, {
      method: 'DELETE',
      headers: this.getHeaders(),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.message || `Failed to delete integration: ${response.statusText}`);
    }
  }

  /**
   * Sync an integration manually
   */
  async syncIntegration(id: string): Promise<{ status: string; message: string }> {
    const response = await fetch(`${this.baseUrl}/integrations/${id}/sync`, {
      method: 'POST',
      headers: this.getHeaders(),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.message || `Failed to sync integration: ${response.statusText}`);
    }

    return response.json();
  }

  /**
   * Test an integration connection
   */
  async testIntegration(id: string): Promise<{ status: string; message: string }> {
    const response = await fetch(`${this.baseUrl}/integrations/${id}/test`, {
      method: 'POST',
      headers: this.getHeaders(),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.message || `Failed to test integration: ${response.statusText}`);
    }

    return response.json();
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
   * Connect to an integration (mainly for OAuth)
   */
  async connectIntegration(id: string): Promise<{ status: string; authorization_url?: string }> {
    const integration = await this.getIntegration(id);

    if (integration.auth_method === 'oauth2') {
      // For OAuth integrations, initiate the OAuth flow
      const result = await this.initiateOAuth(id);
      return {
        status: 'redirect_required',
        authorization_url: result.authorization_url,
      };
    } else {
      // For other auth methods, just update status to connected
      // This might need additional logic based on auth method
      await this.updateIntegration(id, { status: 'connected' });
      return { status: 'connected' };
    }
  }

  /**
   * Disconnect an integration
   */
  async disconnectIntegration(id: string): Promise<void> {
    await this.updateIntegration(id, { status: 'disconnected' });
  }
}

// Export the service class - consumers should instantiate with their config
export const integrationsService = new IntegrationsService();

// Helper function to create configured instance
export function createIntegrationsService(config?: ComponentConfig): IntegrationsService {
  return new IntegrationsService(config);
}
