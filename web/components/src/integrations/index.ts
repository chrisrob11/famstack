/**
 * Integrations Module
 *
 * Role: Central export point for all integration-related components and services
 *
 * Components:
 * - IntegrationGrid: Main grid layout for displaying integrations
 * - IntegrationCard: Individual integration card rendering
 * - IntegrationActions: Action buttons for each integration
 * - AddIntegrationModal: Modal for creating new integrations
 * - OAuthConfigModal: Modal for OAuth provider configuration
 * - CategoryTabs: Navigation tabs for filtering by category
 * - OAuthStatus: OAuth configuration status display
 *
 * Services:
 * - IntegrationsService: Core CRUD operations for integrations
 * - OAuthService: OAuth-specific authentication operations
 * - IntegrationOperations: Business logic for connect/sync/test
 *
 * Types: Integration interfaces and utility functions
 */

export { IntegrationGrid } from './integration-grid.js';
export { IntegrationCard } from './integration-card.js';
export { IntegrationActions } from './integration-actions.js';
export { AddIntegrationModal } from './add-integration-modal.js';
export { OAuthConfigModal } from './oauth-config-modal.js';
export { CategoryTabs } from './category-tabs.js';
export { OAuthStatus } from './oauth-status.js';
export { integrationsService, IntegrationsService } from './integrations-service.js';
export { oAuthService, OAuthService } from './oauth-service.js';
export { integrationOperations, IntegrationOperations } from './integration-operations.js';

export type {
  Integration,
  Provider,
  IntegrationCategory,
  OAuthConfig,
} from './integration-types.js';

export {
  INTEGRATION_CATEGORIES,
  AUTH_DESCRIPTIONS,
  STATUS_LABELS,
  CATEGORY_ICONS,
  getProvidersByCategory,
  getProviderLabel,
  getProviderAuth,
  getCategoryIcon,
} from './integration-types.js';
