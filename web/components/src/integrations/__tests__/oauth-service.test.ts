/**
 * Tests for OAuthService
 * Focuses on OAuth configuration management
 */

import { OAuthService } from '../oauth-service.js';
import { ComponentConfig } from '../../common/types.js';

// Mock fetch globally
global.fetch = jest.fn();

describe('OAuthService', () => {
  let oAuthService: OAuthService;
  let mockConfig: ComponentConfig;

  beforeEach(() => {
    mockConfig = {
      apiBaseUrl: '/api/v1',
      csrfToken: 'test-csrf-token'
    };
    oAuthService = new OAuthService(mockConfig);
    (fetch as jest.Mock).mockClear();
  });

  describe('getOAuthConfig', () => {
    it('should fetch OAuth configuration successfully', async () => {
      const mockConfig = {
        client_id: 'test-client-id',
        configured: true,
        scopes: ['calendar.readonly']
      };

      (fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockConfig
      });

      const result = await oAuthService.getOAuthConfig('google');

      expect(fetch).toHaveBeenCalledWith('/api/v1/config/oauth/google', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': 'test-csrf-token'
        }
      });
      expect(result).toEqual(mockConfig);
    });

    it('should throw error when fetch fails', async () => {
      (fetch as jest.Mock).mockResolvedValueOnce({
        ok: false,
        statusText: 'Not Found'
      });

      await expect(oAuthService.getOAuthConfig('google')).rejects.toThrow(
        'Failed to get OAuth config: Not Found'
      );
    });
  });

  describe('updateOAuthConfig', () => {
    it('should update OAuth configuration', async () => {
      const configUpdate = {
        client_id: 'new-client-id',
        client_secret: 'new-secret',
        configured: true
      };

      const mockUpdatedConfig = {
        ...configUpdate,
        scopes: ['calendar.readonly']
      };

      (fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockUpdatedConfig
      });

      const result = await oAuthService.updateOAuthConfig('google', configUpdate);

      expect(fetch).toHaveBeenCalledWith('/api/v1/config/oauth/google', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': 'test-csrf-token'
        },
        body: JSON.stringify(configUpdate)
      });
      expect(result).toEqual(mockUpdatedConfig);
    });

    it('should handle update errors with error message', async () => {
      (fetch as jest.Mock).mockResolvedValueOnce({
        ok: false,
        json: async () => ({ message: 'Invalid client credentials' })
      });

      await expect(
        oAuthService.updateOAuthConfig('google', { client_id: 'invalid' })
      ).rejects.toThrow('Invalid client credentials');
    });
  });

  describe('initiateOAuth', () => {
    it('should initiate OAuth flow', async () => {
      const mockResponse = {
        authorization_url: 'https://oauth.example.com/auth?client_id=test'
      };

      (fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      });

      const result = await oAuthService.initiateOAuth('integration-123');

      expect(fetch).toHaveBeenCalledWith('/api/v1/integrations/integration-123/oauth/initiate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': 'test-csrf-token'
        }
      });
      expect(result.authorization_url).toBe('https://oauth.example.com/auth?client_id=test');
    });
  });

  describe('revokeOAuth', () => {
    it('should revoke OAuth token', async () => {
      (fetch as jest.Mock).mockResolvedValueOnce({
        ok: true
      });

      await oAuthService.revokeOAuth('integration-123');

      expect(fetch).toHaveBeenCalledWith('/api/v1/integrations/integration-123/oauth/revoke', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': 'test-csrf-token'
        }
      });
    });

    it('should handle revoke errors', async () => {
      (fetch as jest.Mock).mockResolvedValueOnce({
        ok: false,
        json: async () => ({ message: 'Token not found' })
      });

      await expect(oAuthService.revokeOAuth('integration-123')).rejects.toThrow(
        'Token not found'
      );
    });
  });

  describe('validateOAuthConfig', () => {
    it('should validate OAuth configuration', async () => {
      const mockValidation = {
        valid: true,
        errors: []
      };

      (fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockValidation
      });

      const result = await oAuthService.validateOAuthConfig('google');

      expect(fetch).toHaveBeenCalledWith('/api/v1/config/oauth/google/validate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': 'test-csrf-token'
        }
      });
      expect(result.valid).toBe(true);
    });

    it('should return validation errors', async () => {
      const mockValidation = {
        valid: false,
        errors: ['Client ID is required', 'Client secret is invalid']
      };

      (fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockValidation
      });

      const result = await oAuthService.validateOAuthConfig('google');

      expect(result.valid).toBe(false);
      expect(result.errors).toHaveLength(2);
    });
  });
});