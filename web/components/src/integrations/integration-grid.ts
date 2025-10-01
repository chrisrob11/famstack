/**
 * IntegrationGrid Component
 *
 * Role: Displays a grid layout of integration cards and coordinates user actions
 * Responsibilities:
 * - Renders integration cards in a responsive grid layout
 * - Loads and manages integration data from the API
 * - Routes user actions (connect, sync, test) to appropriate services
 * - Dispatches events for parent components to handle
 * - Provides public methods for external integration management
 */

import { LitElement, html, css } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { Integration } from './integration-types.js';
import { integrationsService } from './integrations-service.js';
import { integrationOperations } from './integration-operations.js';
import { errorHandler } from '../common/error-handler.js';
import { EVENTS } from '../common/constants.js';
import './integration-card.js';
import './integration-actions.js';

@customElement('integration-grid')
export class IntegrationGrid extends LitElement {
  @property({ type: String, attribute: 'current-category' })
  currentCategory = 'all';

  @state()
  private integrations: Integration[] = [];

  @state()
  private isLoading = true;

  static override styles = css`
    :host {
      display: block;
    }

    .integrations-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
      gap: 20px;
      margin-top: 20px;
    }

    .loading {
      grid-column: 1 / -1;
      text-align: center;
      padding: 40px;
      color: #6c757d;
      font-size: 16px;
    }

    .empty-state {
      grid-column: 1 / -1;
      text-align: center;
      padding: 40px;
      color: #6c757d;
    }

    .empty-state h3 {
      margin: 0 0 12px 0;
      font-size: 18px;
      color: #495057;
    }

    .empty-state p {
      margin: 0;
      font-size: 14px;
    }
  `;

  override async connectedCallback() {
    super.connectedCallback();
    await this.loadIntegrations();
  }

  override updated(changedProperties: Map<string, any>) {
    if (changedProperties.has('currentCategory')) {
      this.loadIntegrations();
    }
  }

  private async loadIntegrations(): Promise<void> {
    this.isLoading = true;

    const result = await errorHandler.handleAsync(
      async () => {
        const filters = this.currentCategory === 'all' ? {} : { type: this.currentCategory };
        return await integrationsService.getIntegrations(filters);
      },
      { component: 'IntegrationGrid', operation: 'loadIntegrations' },
      []
    );

    this.integrations = result || [];
    this.isLoading = false;
  }

  private async handleIntegrationAction(e: CustomEvent) {
    const { action, integrationId } = e.detail;

    switch (action) {
      case 'configure':
        this.dispatchCustomEvent(EVENTS.CONFIGURE_INTEGRATION, { id: integrationId });
        break;
      case 'delete':
        this.dispatchCustomEvent(EVENTS.DELETE_INTEGRATION, { id: integrationId });
        break;
      case 'connect':
        await this.handleConnect(integrationId);
        break;
      case 'sync':
        await this.handleSync(integrationId);
        break;
      case 'test':
        await this.handleTest(integrationId);
        break;
    }
  }

  private dispatchCustomEvent(type: string, detail: any): void {
    this.dispatchEvent(
      new CustomEvent(type, {
        detail,
        bubbles: true,
      })
    );
  }

  async addIntegration(integrationData: Partial<Integration>): Promise<boolean> {
    const success = await errorHandler.handleAsync(
      async () => {
        await integrationsService.createIntegration(integrationData);
        await this.loadIntegrations();
        return true;
      },
      { component: 'IntegrationGrid', operation: 'addIntegration' }
    );

    return success || false;
  }

  async deleteIntegration(integrationId: string): Promise<boolean> {
    const success = await errorHandler.handleAsync(
      async () => {
        await integrationsService.deleteIntegration(integrationId);
        await this.loadIntegrations();
        return true;
      },
      { component: 'IntegrationGrid', operation: 'deleteIntegration' }
    );

    return success || false;
  }

  private async handleConnect(integrationId: string): Promise<void> {
    const result = await integrationOperations.connectIntegration(integrationId);

    if (result.success) {
      if (result.data?.authorization_url) {
        window.location.href = result.data.authorization_url;
      } else {
        await this.loadIntegrations();
        this.dispatchCustomEvent(EVENTS.INTEGRATION_CONNECTED, { id: integrationId });
      }
    } else {
      errorHandler.dispatchError(
        this,
        EVENTS.INTEGRATION_ERROR,
        new Error(result.message || 'Failed to connect integration'),
        { id: integrationId }
      );
    }
  }

  private async handleSync(integrationId: string): Promise<void> {
    const result = await integrationOperations.syncIntegration(integrationId);

    if (result.success) {
      this.dispatchCustomEvent(EVENTS.INTEGRATION_SYNCED, {
        id: integrationId,
        result: result.data,
      });
      await this.loadIntegrations();
    } else {
      errorHandler.dispatchError(
        this,
        EVENTS.INTEGRATION_ERROR,
        new Error(result.message || 'Failed to sync integration'),
        { id: integrationId }
      );
    }
  }

  private async handleTest(integrationId: string): Promise<void> {
    const result = await integrationOperations.testIntegration(integrationId);

    if (result.success) {
      this.dispatchCustomEvent(EVENTS.INTEGRATION_TESTED, {
        id: integrationId,
        result: result.data,
      });
    } else {
      errorHandler.dispatchError(
        this,
        EVENTS.INTEGRATION_ERROR,
        new Error(result.message || 'Failed to test integration'),
        { id: integrationId }
      );
    }
  }

  setCategory(category: string): void {
    this.currentCategory = category;
  }

  async refresh(): Promise<void> {
    await this.loadIntegrations();
  }

  override render() {
    if (this.isLoading) {
      return html`
        <div class="integrations-grid">
          <div class="loading">Loading integrations...</div>
        </div>
      `;
    }

    if (this.integrations.length === 0) {
      return html`
        <div class="integrations-grid">
          <div class="empty-state">
            <h3>No integrations found</h3>
            <p>Add your first integration to get started connecting external services.</p>
          </div>
        </div>
      `;
    }

    return html`
      <div class="integrations-grid" @integration-action=${this.handleIntegrationAction}>
        ${this.integrations.map(
          integration => html` <integration-card .integration=${integration}></integration-card> `
        )}
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'integration-grid': IntegrationGrid;
  }
}
