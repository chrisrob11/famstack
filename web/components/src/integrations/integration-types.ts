/**
 * IntegrationTypes
 *
 * Role: Type definitions, constants, and utility functions for integrations
 * Responsibilities:
 * - Define TypeScript interfaces for Integration, Provider, OAuthConfig
 * - Export integration categories and provider configurations
 * - Provide utility functions for icons, labels, and provider lookups
 * - Maintain authentication method descriptions
 * - Define status labels and category mappings
 */

export interface Provider {
  value: string;
  label: string;
  auth: string;
}

export interface Integration {
  id: string;
  family_id: string;
  created_by: string;
  integration_type: string;
  provider: string;
  display_name: string;
  description: string;
  status: 'connected' | 'disconnected' | 'error' | 'pending' | 'syncing';
  auth_method: string;
  settings: Record<string, any>; // Parsed from JSON string
  last_sync_at?: string | null;
  last_error?: string | null;
  created_at: string;
  updated_at: string;
}

export interface OAuthConfig {
  client_id: string;
  client_secret: string;
  redirect_url: string;
  scopes: string[];
  configured: boolean;
}

export interface IntegrationCategory {
  key: string;
  label: string;
  icon: string;
  providers: Provider[];
}

export const INTEGRATION_CATEGORIES: IntegrationCategory[] = [
  {
    key: 'calendar',
    label: '📅 Calendar',
    icon: '📅',
    providers: [
      { value: 'google', label: 'Google Calendar', auth: 'oauth2' },
      { value: 'microsoft', label: 'Microsoft Outlook', auth: 'oauth2' },
      { value: 'apple', label: 'Apple Calendar', auth: 'oauth2' },
      { value: 'caldav', label: 'CalDAV', auth: 'basic_auth' },
    ],
  },
  {
    key: 'storage',
    label: '💾 Storage',
    icon: '💾',
    providers: [
      { value: 'google_drive', label: 'Google Drive', auth: 'oauth2' },
      { value: 'dropbox', label: 'Dropbox', auth: 'oauth2' },
      { value: 'onedrive', label: 'Microsoft OneDrive', auth: 'oauth2' },
      { value: 'icloud', label: 'iCloud', auth: 'oauth2' },
    ],
  },
  {
    key: 'communication',
    label: '💬 Communication',
    icon: '💬',
    providers: [
      { value: 'slack', label: 'Slack', auth: 'oauth2' },
      { value: 'discord', label: 'Discord', auth: 'oauth2' },
      { value: 'teams', label: 'Microsoft Teams', auth: 'oauth2' },
      { value: 'email', label: 'Email (SMTP)', auth: 'basic_auth' },
    ],
  },
  {
    key: 'smart_home',
    label: '🏠 Smart Home',
    icon: '🏠',
    providers: [
      { value: 'homekit', label: 'Apple HomeKit', auth: 'api_key' },
      { value: 'alexa', label: 'Amazon Alexa', auth: 'oauth2' },
      { value: 'google_home', label: 'Google Home', auth: 'oauth2' },
    ],
  },
  {
    key: 'automation',
    label: '🤖 Automation',
    icon: '🤖',
    providers: [
      { value: 'ifttt', label: 'IFTTT', auth: 'api_key' },
      { value: 'zapier', label: 'Zapier', auth: 'api_key' },
      { value: 'home_assistant', label: 'Home Assistant', auth: 'api_key' },
    ],
  },
  {
    key: 'finance',
    label: '💰 Finance',
    icon: '💰',
    providers: [
      { value: 'mint', label: 'Mint', auth: 'oauth2' },
      { value: 'ynab', label: 'YNAB', auth: 'api_key' },
      { value: 'plaid', label: 'Plaid', auth: 'api_key' },
    ],
  },
];

export const AUTH_DESCRIPTIONS: Record<string, string> = {
  oauth2:
    "Secure OAuth 2.0 authentication. You'll be redirected to the provider to authorize access.",
  api_key: "API key authentication. You'll need to provide an API key from the provider.",
  basic_auth:
    "Username and password authentication. You'll need to provide your login credentials.",
  webhook: 'Webhook-based integration. The provider will send data to FamStack.',
};

export const STATUS_LABELS: Record<string, string> = {
  connected: 'Connected',
  disconnected: 'Disconnected',
  error: 'Error',
  pending: 'Pending',
  syncing: 'Syncing',
};

export const CATEGORY_ICONS: Record<string, string> = {
  calendar: '📅',
  storage: '💾',
  communication: '💬',
  smart_home: '🏠',
  automation: '🤖',
  finance: '💰',
};

// Helper functions
export function getProvidersByCategory(category: string): Provider[] {
  const categoryData = INTEGRATION_CATEGORIES.find(c => c.key === category);
  return categoryData ? categoryData.providers : [];
}

export function getProviderLabel(provider: string): string {
  for (const category of INTEGRATION_CATEGORIES) {
    const found = category.providers.find(p => p.value === provider);
    if (found) return found.label;
  }
  return provider;
}

export function getProviderAuth(provider: string): string {
  for (const category of INTEGRATION_CATEGORIES) {
    const found = category.providers.find(p => p.value === provider);
    if (found) return found.auth;
  }
  return 'oauth2';
}

export function getCategoryIcon(type: string): string {
  return CATEGORY_ICONS[type] || '🔗';
}
