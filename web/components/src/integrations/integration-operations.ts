/**
 * IntegrationOperations Service
 *
 * Role: Business logic service for integration state operations
 * Responsibilities:
 * - Handle integration connection logic (OAuth vs API key vs basic auth)
 * - Coordinate between IntegrationsService and OAuthService
 * - Manage integration disconnect and cleanup operations
 * - Provide unified interface for sync and test operations
 * - Handle bulk operations across multiple integrations
 * - Return consistent OperationResult objects with success/error states
 */

import { integrationsService } from './integrations-service.js';
import { oAuthService } from './oauth-service.js';
import { ComponentConfig } from '../common/types.js';
import { AUTH_METHODS } from '../common/constants.js';

export interface OperationResult {
  success: boolean;
  message?: string;
  data?: any;
}

export class IntegrationOperations {
  constructor(_config?: ComponentConfig) {
    // Config parameter available for future use
  }

  /**
   * Connect an integration based on its authentication method
   */
  async connectIntegration(integrationId: string): Promise<OperationResult> {
    try {
      const integration = await integrationsService.getIntegration(integrationId);

      if (integration.auth_method === AUTH_METHODS.OAUTH2) {
        // For OAuth integrations, initiate the OAuth flow
        const result = await oAuthService.initiateOAuth(integrationId);
        return {
          success: true,
          message: 'OAuth flow initiated',
          data: { authorization_url: result.authorization_url },
        };
      } else if (integration.auth_method === AUTH_METHODS.API_KEY) {
        // For API key integrations, just update status to connected
        await integrationsService.updateIntegration(integrationId, { status: 'connected' });
        return {
          success: true,
          message: 'Integration connected successfully',
        };
      } else {
        // For basic auth or other methods
        await integrationsService.updateIntegration(integrationId, { status: 'connected' });
        return {
          success: true,
          message: 'Integration connected successfully',
        };
      }
    } catch (error) {
      return {
        success: false,
        message: error instanceof Error ? error.message : 'Failed to connect integration',
      };
    }
  }

  /**
   * Disconnect an integration
   */
  async disconnectIntegration(integrationId: string): Promise<OperationResult> {
    try {
      const integration = await integrationsService.getIntegration(integrationId);

      if (integration.auth_method === AUTH_METHODS.OAUTH2) {
        // For OAuth integrations, revoke the token
        await oAuthService.revokeOAuth(integrationId);
      }

      // Update status to disconnected
      await integrationsService.updateIntegration(integrationId, { status: 'disconnected' });

      return {
        success: true,
        message: 'Integration disconnected successfully',
      };
    } catch (error) {
      return {
        success: false,
        message: error instanceof Error ? error.message : 'Failed to disconnect integration',
      };
    }
  }

  /**
   * Sync an integration
   */
  async syncIntegration(integrationId: string): Promise<OperationResult> {
    try {
      const result = await integrationsService.syncIntegration(integrationId);
      return {
        success: true,
        message: result.message || 'Integration synced successfully',
        data: result,
      };
    } catch (error) {
      return {
        success: false,
        message: error instanceof Error ? error.message : 'Failed to sync integration',
      };
    }
  }

  /**
   * Test an integration connection
   */
  async testIntegration(integrationId: string): Promise<OperationResult> {
    try {
      const result = await integrationsService.testIntegration(integrationId);
      return {
        success: true,
        message: result.message || 'Integration test completed successfully',
        data: result,
      };
    } catch (error) {
      return {
        success: false,
        message: error instanceof Error ? error.message : 'Failed to test integration',
      };
    }
  }

  /**
   * Refresh OAuth token for an integration
   */
  async refreshIntegrationToken(integrationId: string): Promise<OperationResult> {
    try {
      const integration = await integrationsService.getIntegration(integrationId);

      if (integration.auth_method !== AUTH_METHODS.OAUTH2) {
        return {
          success: false,
          message: 'Token refresh is only available for OAuth integrations',
        };
      }

      const result = await oAuthService.refreshOAuth(integrationId);
      return {
        success: result.success,
        message: result.success ? 'Token refreshed successfully' : 'Failed to refresh token',
      };
    } catch (error) {
      return {
        success: false,
        message: error instanceof Error ? error.message : 'Failed to refresh integration token',
      };
    }
  }

  /**
   * Perform a bulk operation on multiple integrations
   */
  async bulkOperation(
    integrationIds: string[],
    operation: 'connect' | 'disconnect' | 'sync' | 'test'
  ): Promise<{
    results: Record<string, OperationResult>;
    summary: { success: number; failed: number };
  }> {
    const results: Record<string, OperationResult> = {};
    let successCount = 0;
    let failedCount = 0;

    for (const id of integrationIds) {
      let result: OperationResult;

      switch (operation) {
        case 'connect':
          result = await this.connectIntegration(id);
          break;
        case 'disconnect':
          result = await this.disconnectIntegration(id);
          break;
        case 'sync':
          result = await this.syncIntegration(id);
          break;
        case 'test':
          result = await this.testIntegration(id);
          break;
        default:
          result = { success: false, message: 'Unknown operation' };
      }

      results[id] = result;
      if (result.success) {
        successCount++;
      } else {
        failedCount++;
      }
    }

    return {
      results,
      summary: { success: successCount, failed: failedCount },
    };
  }
}

// Export singleton instance
export const integrationOperations = new IntegrationOperations();

// Helper function to create configured instance
export function createIntegrationOperations(config?: ComponentConfig): IntegrationOperations {
  return new IntegrationOperations(config);
}
