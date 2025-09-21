/**
 * Integration Grid component
 * Displays and manages the list of integrations
 */

import { ComponentConfig } from '../common/types.js';
import {
  Integration,
  getCategoryIcon,
  getProviderLabel,
  STATUS_LABELS,
} from './integration-types.js';
import { integrationsService } from './integrations-service.js';
import { loadCSS } from '../common/dom-utils.js';
import { logger } from '../common/logger.js';

export class IntegrationGrid {
  private container: HTMLElement;
  private config: ComponentConfig;
  private integrations: Integration[] = [];
  private currentCategory: string = 'all';

  constructor(container: HTMLElement, config: ComponentConfig) {
    this.container = container;
    this.config = config;
  }

  async init(): Promise<void> {
    await this.loadStyles();
    this.render();
    this.setupEventListeners();
    await this.loadIntegrations();
  }

  private async loadStyles(): Promise<void> {
    try {
      await loadCSS(
        '/components/src/integrations/styles/integration-grid.css',
        'integration-grid-styles'
      );
    } catch (error) {
      logger.styleError('integration grid', error);
    }
  }

  private render(): void {
    this.container.innerHTML = `
      <div class="integrations-grid">
        <div class="loading">Loading integrations...</div>
      </div>
    `;
  }

  private setupEventListeners(): void {
    this.container.addEventListener('click', e => {
      const target = e.target as HTMLElement;
      const card = target.closest('.integration-card');

      if (!card) return;

      const integrationId = card.getAttribute('data-integration-id');
      if (!integrationId) return;

      if (target.classList.contains('configure-btn')) {
        this.dispatchEvent('configure-integration', { id: integrationId });
      } else if (target.classList.contains('delete-btn')) {
        this.dispatchEvent('delete-integration', { id: integrationId });
      } else if (target.classList.contains('connect-btn')) {
        this.handleConnect(integrationId);
      } else if (target.classList.contains('sync-btn')) {
        this.handleSync(integrationId);
      } else if (target.classList.contains('test-btn')) {
        this.handleTest(integrationId);
      }
    });
  }

  private dispatchEvent(type: string, detail: any): void {
    this.container.dispatchEvent(
      new CustomEvent(type, {
        detail,
        bubbles: true,
      })
    );
  }

  async loadIntegrations(category: string = this.currentCategory): Promise<void> {
    this.currentCategory = category;

    try {
      const filters = category === 'all' ? {} : { type: category };
      this.integrations = await integrationsService.getIntegrations(filters);
    } catch (error) {
      logger.loadError('integrations', error);
      this.integrations = [];
    }

    this.renderIntegrations();
  }

  private renderIntegrations(): void {
    const gridContainer = this.container.querySelector('.integrations-grid');
    if (!gridContainer) return;

    if (this.integrations.length === 0) {
      gridContainer.innerHTML = `
        <div class="empty-state">
          <h3>No integrations found</h3>
          <p>Add your first integration to get started connecting external services.</p>
        </div>
      `;
      return;
    }

    gridContainer.innerHTML = this.integrations
      .map(
        integration => `
      <div class="integration-card" data-integration-id="${integration.id}">
        <div class="integration-header">
          <div class="integration-icon ${integration.integration_type}">
            ${getCategoryIcon(integration.integration_type)}
          </div>
          <div class="integration-info">
            <h3>${integration.display_name}</h3>
            <p class="provider">${getProviderLabel(integration.provider)}</p>
          </div>
        </div>
        <div class="integration-status">
          <div class="status-indicator ${integration.status}"></div>
          <span>${STATUS_LABELS[integration.status] || integration.status}</span>
        </div>
        ${integration.description ? `<p class="integration-description">${integration.description}</p>` : ''}
        <div class="integration-actions">
          ${this.renderIntegrationActions(integration)}
        </div>
      </div>
    `
      )
      .join('');
  }

  private renderIntegrationActions(integration: Integration): string {
    const actions = [];

    if (
      integration.status === 'disconnected' ||
      (integration.status === 'pending' && integration.auth_method === 'oauth2')
    ) {
      actions.push(`<button class="btn btn-secondary connect-btn">Connect</button>`);
    } else if (integration.status === 'connected') {
      actions.push(`<button class="btn btn-secondary sync-btn">Sync</button>`);
      actions.push(`<button class="btn btn-secondary test-btn">Test</button>`);
    }

    actions.push(`<button class="btn btn-secondary configure-btn">Configure</button>`);
    actions.push(`<button class="btn btn-danger delete-btn">Delete</button>`);

    return actions.join('');
  }

  async addIntegration(integrationData: Partial<Integration>): Promise<boolean> {
    try {
      await integrationsService.createIntegration(integrationData);
      await this.loadIntegrations();
      return true;
    } catch (error) {
      logger.error('Error creating integration:', error);
      throw error;
    }
  }

  async deleteIntegration(integrationId: string): Promise<boolean> {
    try {
      await integrationsService.deleteIntegration(integrationId);
      await this.loadIntegrations();
      return true;
    } catch (error) {
      logger.error('Error deleting integration:', error);
      throw error;
    }
  }

  async handleConnect(integrationId: string): Promise<void> {
    try {
      const result = await integrationsService.connectIntegration(integrationId);

      if (result.authorization_url) {
        // Redirect to OAuth authorization URL
        window.location.href = result.authorization_url;
      } else {
        // Integration connected successfully
        await this.loadIntegrations();
        this.dispatchEvent('integration-connected', { id: integrationId });
      }
    } catch (error) {
      logger.error('Error connecting integration:', error);
      this.dispatchEvent('integration-error', {
        id: integrationId,
        error: error instanceof Error ? error.message : 'Failed to connect integration',
      });
    }
  }

  async handleSync(integrationId: string): Promise<void> {
    try {
      const result = await integrationsService.syncIntegration(integrationId);
      this.dispatchEvent('integration-synced', { id: integrationId, result });
      // Reload to get updated status
      await this.loadIntegrations();
    } catch (error) {
      logger.error('Error syncing integration:', error);
      this.dispatchEvent('integration-error', {
        id: integrationId,
        error: error instanceof Error ? error.message : 'Failed to sync integration',
      });
    }
  }

  async handleTest(integrationId: string): Promise<void> {
    try {
      const result = await integrationsService.testIntegration(integrationId);
      this.dispatchEvent('integration-tested', { id: integrationId, result });
    } catch (error) {
      logger.error('Error testing integration:', error);
      this.dispatchEvent('integration-error', {
        id: integrationId,
        error: error instanceof Error ? error.message : 'Failed to test integration',
      });
    }
  }

  setCategory(category: string): void {
    this.loadIntegrations(category);
  }

  async refresh(): Promise<void> {
    await this.loadIntegrations();
  }

  destroy(): void {
    // Component cleanup
  }
}
