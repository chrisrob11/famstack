/**
 * IntegrationDetailsModal Component
 *
 * Role: Display detailed information about an integration in a modal
 * Responsibilities:
 * - Show integration details, credentials, and sync history
 * - Handle OAuth connection flow within modal (popup)
 * - Provide actions for test, sync, and reconnect
 * - Display sync history and logs
 */

import { LitElement, html, css } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { Integration } from './integration-types.js';
import { integrationsService } from './integrations-service.js';
import { integrationOperations } from './integration-operations.js';
import { modalStyles, buttonStyles } from '../common/shared-styles.js';
import { errorHandler } from '../common/error-handler.js';
import { EVENTS } from '../common/constants.js';

interface SyncHistory {
  id: string;
  status: string;
  sync_type: string;
  started_at: string;
  completed_at?: string;
  items_synced?: number;
  error_message?: string;
}

@customElement('integration-details-modal')
export class IntegrationDetailsModal extends LitElement {
  @property({ type: Boolean })
  open = false;

  @property({ type: String })
  integrationId = '';

  @state()
  private integration: Integration | null = null;

  @state()
  private syncHistory: SyncHistory[] = [];

  @state()
  private isLoading = true;

  @state()
  private actionInProgress = false;

  static override styles = [
    buttonStyles,
    modalStyles,
    css`
      :host {
        display: none;
      }

      :host([open]) {
        display: block;
      }

      .details-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
        gap: 16px;
        margin-bottom: 24px;
      }

      .detail-item {
        display: flex;
        flex-direction: column;
        gap: 4px;
      }

      .detail-item label {
        font-weight: 600;
        color: #495057;
        font-size: 12px;
        text-transform: uppercase;
        letter-spacing: 0.5px;
      }

      .detail-item span {
        color: #333;
        font-size: 14px;
      }

      .status-badge {
        display: inline-block;
        padding: 4px 8px;
        border-radius: 4px;
        font-size: 12px;
        font-weight: 500;
        text-transform: capitalize;
      }

      .status-badge.connected {
        background: #d4edda;
        color: #155724;
      }

      .status-badge.disconnected {
        background: #f8d7da;
        color: #721c24;
      }

      .status-badge.error {
        background: #fff3cd;
        color: #856404;
      }

      .status-badge.pending {
        background: #e2e3e5;
        color: #383d41;
      }

      .detail-section {
        margin-bottom: 24px;
      }

      .detail-section h4 {
        margin: 0 0 12px 0;
        color: #333;
        font-size: 16px;
        font-weight: 600;
      }

      .detail-section p {
        margin: 0;
        color: #6c757d;
        font-size: 14px;
        line-height: 1.5;
      }

      .sync-history {
        display: flex;
        flex-direction: column;
        gap: 8px;
      }

      .sync-item {
        display: flex;
        align-items: center;
        gap: 12px;
        padding: 12px;
        background: #f8f9fa;
        border-radius: 4px;
        font-size: 12px;
      }

      .sync-status {
        padding: 2px 8px;
        border-radius: 4px;
        font-weight: 500;
        text-transform: capitalize;
      }

      .sync-status.success {
        background: #d4edda;
        color: #155724;
      }

      .sync-status.error {
        background: #f8d7da;
        color: #721c24;
      }

      .sync-status.partial {
        background: #fff3cd;
        color: #856404;
      }

      .sync-type {
        color: #6c757d;
      }

      .sync-time {
        color: #adb5bd;
        margin-left: auto;
      }

      .sync-count {
        color: #6c757d;
        background: #e9ecef;
        padding: 2px 8px;
        border-radius: 4px;
      }

      .loading-spinner {
        text-align: center;
        padding: 40px;
        color: #6c757d;
      }

      .action-buttons {
        display: flex;
        gap: 8px;
        flex-wrap: wrap;
        margin-bottom: 24px;
      }

      .info-box {
        background: #e7f3ff;
        border: 1px solid #b3d9ff;
        border-radius: 4px;
        padding: 12px;
        margin-bottom: 16px;
      }

      .info-box p {
        margin: 0;
        color: #004085;
        font-size: 14px;
      }

      .oauth-popup-message {
        background: #fff3cd;
        border: 1px solid #ffc107;
        border-radius: 4px;
        padding: 12px;
        margin: 12px 0;
        font-size: 14px;
        color: #856404;
      }
    `,
  ];

  override updated(changedProperties: Map<string, any>) {
    if (changedProperties.has('open') && this.open) {
      this.loadIntegrationDetails();
    }
  }

  private async loadIntegrationDetails(): Promise<void> {
    if (!this.integrationId) return;

    this.isLoading = true;

    const result = await errorHandler.handleAsync(
      async () => {
        return await integrationsService.getIntegrationDetails(this.integrationId);
      },
      { component: 'IntegrationDetailsModal', operation: 'loadDetails' }
    );

    if (result) {
      this.integration = result.integration;
      this.syncHistory = result.recent_sync_history || [];
    }

    this.isLoading = false;
  }

  private close() {
    this.open = false;
    this.dispatchEvent(new CustomEvent('close', { bubbles: true }));
  }

  private async handleConnect() {
    if (!this.integration) return;

    this.actionInProgress = true;

    const result = await integrationOperations.connectIntegration(this.integration.id);

    if (result.success && result.data?.authorization_url) {
      // Open OAuth in popup window
      const popup = window.open(
        result.data.authorization_url,
        'oauth',
        'width=600,height=700,scrollbars=yes'
      );

      // Poll for popup close or listen for message
      const checkPopup = setInterval(() => {
        if (popup?.closed) {
          clearInterval(checkPopup);
          this.actionInProgress = false;
          this.loadIntegrationDetails();
          this.dispatchEvent(
            new CustomEvent(EVENTS.INTEGRATION_CONNECTED, {
              detail: { id: this.integration!.id },
              bubbles: true,
            })
          );
        }
      }, 1000);
    } else if (result.success) {
      // Non-OAuth connection succeeded
      this.actionInProgress = false;
      await this.loadIntegrationDetails();
      this.dispatchEvent(
        new CustomEvent(EVENTS.INTEGRATION_CONNECTED, {
          detail: { id: this.integration.id },
          bubbles: true,
        })
      );
    } else {
      this.actionInProgress = false;
      errorHandler.dispatchError(
        this,
        EVENTS.INTEGRATION_ERROR,
        new Error(result.message || 'Failed to connect'),
        { id: this.integration.id }
      );
    }
  }

  private async handleSync() {
    if (!this.integration) return;

    this.actionInProgress = true;

    const result = await integrationOperations.syncIntegration(this.integration.id);

    this.actionInProgress = false;

    if (result.success) {
      await this.loadIntegrationDetails();
      this.dispatchEvent(
        new CustomEvent(EVENTS.INTEGRATION_SYNCED, {
          detail: { id: this.integration.id, result: result.data },
          bubbles: true,
        })
      );
    } else {
      errorHandler.dispatchError(
        this,
        EVENTS.INTEGRATION_ERROR,
        new Error(result.message || 'Failed to sync'),
        { id: this.integration.id }
      );
    }
  }

  private async handleTest() {
    if (!this.integration) return;

    this.actionInProgress = true;

    const result = await integrationOperations.testIntegration(this.integration.id);

    this.actionInProgress = false;

    if (result.success) {
      this.dispatchEvent(
        new CustomEvent(EVENTS.INTEGRATION_TESTED, {
          detail: { id: this.integration.id, result: result.data },
          bubbles: true,
        })
      );
      alert('Integration test successful!');
    } else {
      errorHandler.dispatchError(
        this,
        EVENTS.INTEGRATION_ERROR,
        new Error(result.message || 'Failed to test'),
        { id: this.integration.id }
      );
    }
  }

  show(integrationId: string) {
    this.integrationId = integrationId;
    this.open = true;
  }

  hide() {
    this.close();
  }

  private renderLoading() {
    return html` <div class="loading-spinner">Loading integration details...</div> `;
  }

  private renderDetails() {
    if (!this.integration) return html``;

    const needsConnection =
      this.integration.status === 'pending' || this.integration.status === 'error';

    return html`
      ${needsConnection
        ? html`
            <div class="info-box">
              <p>
                This integration needs to be connected. Click "Connect" to
                ${this.integration.auth_method === 'oauth2'
                  ? 'authorize with OAuth'
                  : 'complete the setup'}.
              </p>
            </div>
          `
        : ''}

      <div class="action-buttons">
        ${needsConnection
          ? html`
              <button
                class="btn btn-primary"
                @click=${this.handleConnect}
                ?disabled=${this.actionInProgress}
              >
                ${this.integration.auth_method === 'oauth2' ? 'Connect with OAuth' : 'Connect'}
              </button>
            `
          : html`
              <button
                class="btn btn-secondary btn-sm"
                @click=${this.handleSync}
                ?disabled=${this.actionInProgress}
              >
                Sync Now
              </button>
              <button
                class="btn btn-secondary btn-sm"
                @click=${this.handleTest}
                ?disabled=${this.actionInProgress}
              >
                Test Connection
              </button>
              <button
                class="btn btn-secondary btn-sm"
                @click=${this.handleConnect}
                ?disabled=${this.actionInProgress}
              >
                Reconnect
              </button>
            `}
      </div>

      ${this.actionInProgress && this.integration.auth_method === 'oauth2'
        ? html`
            <div class="oauth-popup-message">
              Please complete the authorization in the popup window. If you don't see a popup, check
              if your browser is blocking popups.
            </div>
          `
        : ''}

      <div class="details-grid">
        <div class="detail-item">
          <label>Provider</label>
          <span>${this.integration.provider}</span>
        </div>
        <div class="detail-item">
          <label>Type</label>
          <span>${this.integration.integration_type}</span>
        </div>
        <div class="detail-item">
          <label>Status</label>
          <span class="status-badge ${this.integration.status}">${this.integration.status}</span>
        </div>
        <div class="detail-item">
          <label>Authentication</label>
          <span>${this.integration.auth_method}</span>
        </div>
        <div class="detail-item">
          <label>Created</label>
          <span>${new Date(this.integration.created_at).toLocaleDateString()}</span>
        </div>
        ${this.integration.last_sync_at
          ? html`
              <div class="detail-item">
                <label>Last Sync</label>
                <span>${new Date(this.integration.last_sync_at).toLocaleString()}</span>
              </div>
            `
          : ''}
      </div>

      ${this.integration.description
        ? html`
            <div class="detail-section">
              <h4>Description</h4>
              <p>${this.integration.description}</p>
            </div>
          `
        : ''}
      ${this.syncHistory.length > 0
        ? html`
            <div class="detail-section">
              <h4>Recent Sync History</h4>
              <div class="sync-history">
                ${this.syncHistory.map(
                  sync => html`
                    <div class="sync-item">
                      <span class="sync-status ${sync.status}">${sync.status}</span>
                      <span class="sync-type">${sync.sync_type}</span>
                      ${sync.items_synced
                        ? html`<span class="sync-count">${sync.items_synced} items</span>`
                        : ''}
                      <span class="sync-time">${new Date(sync.started_at).toLocaleString()}</span>
                    </div>
                  `
                )}
              </div>
            </div>
          `
        : ''}
    `;
  }

  override render() {
    if (!this.open) return html``;

    return html`
      <div class="modal" @click=${(e: Event) => e.target === e.currentTarget && this.close()}>
        <div class="modal-content">
          <div class="modal-header">
            <h2>${this.integration?.display_name || 'Integration Details'}</h2>
            <button class="close-btn" @click=${this.close}>&times;</button>
          </div>
          <div class="modal-body">
            ${this.isLoading ? this.renderLoading() : this.renderDetails()}
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click=${this.close}>Close</button>
          </div>
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'integration-details-modal': IntegrationDetailsModal;
  }
}
