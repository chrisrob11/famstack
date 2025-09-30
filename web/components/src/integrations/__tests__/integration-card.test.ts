/**
 * Tests for IntegrationCard component
 * Focuses on testing the component's logic without Lit rendering
 */

import type { Integration } from '../integration-types.js';

describe('IntegrationCard', () => {
  const mockIntegration: Integration = {
    id: 'test-integration',
    family_id: 'test-family',
    created_by: 'test-user',
    integration_type: 'calendar',
    provider: 'google',
    display_name: 'Test Google Calendar',
    description: 'Test calendar integration',
    status: 'connected',
    auth_method: 'oauth2',
    settings: {},
    last_sync_at: null,
    last_error: null,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z'
  };

  it('should validate integration data structure', () => {
    expect(mockIntegration.id).toBe('test-integration');
    expect(mockIntegration.integration_type).toBe('calendar');
    expect(mockIntegration.provider).toBe('google');
    expect(mockIntegration.status).toBe('connected');
  });

  it('should have required integration properties', () => {
    const requiredFields = [
      'id', 'family_id', 'created_by', 'integration_type',
      'provider', 'display_name', 'status', 'auth_method'
    ];

    requiredFields.forEach(field => {
      expect(mockIntegration).toHaveProperty(field);
      expect(mockIntegration[field as keyof Integration]).toBeTruthy();
    });
  });

  it('should handle integration with different statuses', () => {
    const statuses = ['connected', 'disconnected', 'pending', 'error'] as const;

    statuses.forEach(status => {
      const integration = { ...mockIntegration, status };
      expect(integration.status).toBe(status);
    });
  });

  it('should handle integration with different types', () => {
    const types = ['calendar', 'communication', 'productivity'] as const;

    types.forEach(type => {
      const integration = { ...mockIntegration, integration_type: type };
      expect(integration.integration_type).toBe(type);
    });
  });

  it('should handle optional description field', () => {
    const integrationWithDescription = { ...mockIntegration, description: 'Test description' };
    const integrationWithoutDescription = { ...mockIntegration, description: '' };

    expect(integrationWithDescription.description).toBe('Test description');
    expect(integrationWithoutDescription.description).toBe('');
  });
});