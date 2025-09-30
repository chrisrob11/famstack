/**
 * Tests for IntegrationOperations service
 * Focuses on critical business logic for connect/sync/test operations
 */

import { IntegrationOperations, OperationResult } from '../integration-operations.js';
import { integrationsService } from '../integrations-service.js';
import { oAuthService } from '../oauth-service.js';
import { AUTH_METHODS } from '../../common/constants.js';

// Mock the services
jest.mock('../integrations-service.js');
jest.mock('../oauth-service.js');

const mockIntegrationsService = integrationsService as jest.Mocked<typeof integrationsService>;
const mockOAuthService = oAuthService as jest.Mocked<typeof oAuthService>;

describe('IntegrationOperations', () => {
  let operations: IntegrationOperations;

  beforeEach(() => {
    operations = new IntegrationOperations();
    jest.clearAllMocks();
  });

  describe('connectIntegration', () => {
    it('should handle OAuth2 integration connection', async () => {
      const mockIntegration = {
        id: 'test-id',
        auth_method: AUTH_METHODS.OAUTH2,
        provider: 'google',
        integration_type: 'calendar'
      };

      const mockOAuthResult = {
        authorization_url: 'https://oauth.example.com/auth'
      };

      mockIntegrationsService.getIntegration.mockResolvedValue(mockIntegration as any);
      mockOAuthService.initiateOAuth.mockResolvedValue(mockOAuthResult);

      const result = await operations.connectIntegration('test-id');

      expect(result.success).toBe(true);
      expect(result.data?.authorization_url).toBe('https://oauth.example.com/auth');
      expect(mockOAuthService.initiateOAuth).toHaveBeenCalledWith('test-id');
    });

    it('should handle API key integration connection', async () => {
      const mockIntegration = {
        id: 'test-id',
        auth_method: AUTH_METHODS.API_KEY,
        provider: 'custom',
        integration_type: 'productivity'
      };

      mockIntegrationsService.getIntegration.mockResolvedValue(mockIntegration as any);
      mockIntegrationsService.updateIntegration.mockResolvedValue(mockIntegration as any);

      const result = await operations.connectIntegration('test-id');

      expect(result.success).toBe(true);
      expect(result.message).toContain('connected successfully');
      expect(mockIntegrationsService.updateIntegration).toHaveBeenCalledWith('test-id', { status: 'connected' });
    });

    it('should handle connection errors gracefully', async () => {
      mockIntegrationsService.getIntegration.mockRejectedValue(new Error('Integration not found'));

      const result = await operations.connectIntegration('invalid-id');

      expect(result.success).toBe(false);
      expect(result.message).toContain('Integration not found');
    });
  });

  describe('syncIntegration', () => {
    it('should sync integration successfully', async () => {
      const mockSyncResult = {
        status: 'success',
        message: 'Synced 5 calendar events'
      };

      mockIntegrationsService.syncIntegration.mockResolvedValue(mockSyncResult);

      const result = await operations.syncIntegration('test-id');

      expect(result.success).toBe(true);
      expect(result.message).toBe('Synced 5 calendar events');
      expect(result.data).toEqual(mockSyncResult);
    });

    it('should handle sync errors', async () => {
      mockIntegrationsService.syncIntegration.mockRejectedValue(new Error('Sync failed'));

      const result = await operations.syncIntegration('test-id');

      expect(result.success).toBe(false);
      expect(result.message).toContain('Sync failed');
    });
  });

  describe('testIntegration', () => {
    it('should test integration successfully', async () => {
      const mockTestResult = {
        status: 'success',
        message: 'Connection verified'
      };

      mockIntegrationsService.testIntegration.mockResolvedValue(mockTestResult);

      const result = await operations.testIntegration('test-id');

      expect(result.success).toBe(true);
      expect(result.message).toBe('Connection verified');
    });
  });

  describe('disconnectIntegration', () => {
    it('should disconnect OAuth integration with token revocation', async () => {
      const mockIntegration = {
        id: 'test-id',
        auth_method: AUTH_METHODS.OAUTH2
      };

      mockIntegrationsService.getIntegration.mockResolvedValue(mockIntegration as any);
      mockOAuthService.revokeOAuth.mockResolvedValue();
      mockIntegrationsService.updateIntegration.mockResolvedValue(mockIntegration as any);

      const result = await operations.disconnectIntegration('test-id');

      expect(result.success).toBe(true);
      expect(mockOAuthService.revokeOAuth).toHaveBeenCalledWith('test-id');
      expect(mockIntegrationsService.updateIntegration).toHaveBeenCalledWith('test-id', { status: 'disconnected' });
    });

    it('should disconnect non-OAuth integration', async () => {
      const mockIntegration = {
        id: 'test-id',
        auth_method: AUTH_METHODS.API_KEY
      };

      mockIntegrationsService.getIntegration.mockResolvedValue(mockIntegration as any);
      mockIntegrationsService.updateIntegration.mockResolvedValue(mockIntegration as any);

      const result = await operations.disconnectIntegration('test-id');

      expect(result.success).toBe(true);
      expect(mockOAuthService.revokeOAuth).not.toHaveBeenCalled();
    });
  });

  describe('bulkOperation', () => {
    it('should perform bulk sync operation', async () => {
      const integrationIds = ['id1', 'id2', 'id3'];

      mockIntegrationsService.syncIntegration
        .mockResolvedValueOnce({ status: 'success', message: 'Synced' })
        .mockResolvedValueOnce({ status: 'success', message: 'Synced' })
        .mockRejectedValueOnce(new Error('Sync failed'));

      const result = await operations.bulkOperation(integrationIds, 'sync');

      expect(result.summary.success).toBe(2);
      expect(result.summary.failed).toBe(1);
      expect(result.results['id1'].success).toBe(true);
      expect(result.results['id2'].success).toBe(true);
      expect(result.results['id3'].success).toBe(false);
    });
  });
});