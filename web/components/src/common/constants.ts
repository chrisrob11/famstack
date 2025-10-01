// Integration Status Constants
export const INTEGRATION_STATUS = {
  CONNECTED: 'connected',
  DISCONNECTED: 'disconnected',
  PENDING: 'pending',
  ERROR: 'error',
} as const;

export type IntegrationStatus = (typeof INTEGRATION_STATUS)[keyof typeof INTEGRATION_STATUS];

// OAuth Status Constants
export const OAUTH_STATUS = {
  CONFIGURED: 'configured',
  NOT_CONFIGURED: 'not-configured',
  LOADING: 'loading',
  ERROR: 'error',
} as const;

export type OAuthStatus = (typeof OAUTH_STATUS)[keyof typeof OAUTH_STATUS];

// API Endpoints Configuration
export const API_ENDPOINTS = {
  OAUTH_GOOGLE_CONFIG: '/api/v1/config/oauth/google',
  INTEGRATIONS: '/api/v1/integrations',
  OAUTH_CALLBACK_GOOGLE: '/oauth/google/callback',
} as const;

// Authentication Methods
export const AUTH_METHODS = {
  OAUTH2: 'oauth2',
  API_KEY: 'api_key',
  BASIC: 'basic',
} as const;

export type AuthMethod = (typeof AUTH_METHODS)[keyof typeof AUTH_METHODS];

// Event Names
export const EVENTS = {
  CATEGORY_CHANGED: 'category-changed',
  CREATE_INTEGRATION: 'create-integration',
  DELETE_INTEGRATION: 'delete-integration',
  CONFIGURE_INTEGRATION: 'configure-integration',
  INTEGRATION_CONNECTED: 'integration-connected',
  INTEGRATION_SYNCED: 'integration-synced',
  INTEGRATION_TESTED: 'integration-tested',
  INTEGRATION_ERROR: 'integration-error',
  OAUTH_UPDATED: 'oauth-updated',
} as const;
