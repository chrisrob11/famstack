/**
 * Integrations module exports
 */

export { IntegrationGrid } from './integration-grid.js';
export { AddIntegrationModal } from './add-integration-modal.js';
export { OAuthConfigModal } from './oauth-config-modal.js';
export { CategoryTabs } from './category-tabs.js';
export { OAuthStatus } from './oauth-status.js';
export { integrationsService, IntegrationsService } from './integrations-service.js';

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
