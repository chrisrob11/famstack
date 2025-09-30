/**
 * OAuthStatus Component
 *
 * Role: Displays the current OAuth configuration status
 * Responsibilities:
 * - Show overall OAuth configuration status badge
 * - Display per-provider configuration status (Google, etc.)
 * - Load OAuth status from backend API
 * - Provide refresh capability for status updates
 * - Handle loading and error states gracefully
 */

import { LitElement, html, css } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { errorHandler } from '../common/error-handler.js';
import { API_ENDPOINTS, OAUTH_STATUS } from '../common/constants.js';

@customElement('oauth-status')
export class OAuthStatus extends LitElement {
  @state()
  private googleConfigured = false;

  @state()
  private isLoading = true;

  @state()
  private hasError = false;

  static override styles = css`
    :host {
      display: block;
    }

    .oauth-status-section {
      margin-bottom: 24px;
    }

    .status-card {
      background: #fff;
      border: 1px solid #e1e5e9;
      border-radius: 8px;
      padding: 20px;
      box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    }

    .status-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 12px;
    }

    .status-header h3 {
      margin: 0;
      font-size: 18px;
      font-weight: 600;
      color: #333;
    }

    .status-badge {
      padding: 4px 12px;
      border-radius: 16px;
      font-size: 12px;
      font-weight: 600;
      text-transform: uppercase;
    }

    .status-badge.loading {
      background: #f8f9fa;
      color: #6c757d;
    }

    .status-badge.configured {
      background: #d4edda;
      color: #155724;
    }

    .status-badge.not-configured {
      background: #f8d7da;
      color: #721c24;
    }

    .status-description {
      margin: 0 0 16px 0;
      color: #6c757d;
      font-size: 14px;
    }

    .oauth-providers {
      display: flex;
      flex-direction: column;
      gap: 8px;
    }

    .provider-status {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 8px 12px;
      border-radius: 6px;
      font-size: 14px;
    }

    .provider-status.configured {
      background: #d4edda;
      color: #155724;
    }

    .provider-status.not-configured {
      background: #f8d7da;
      color: #721c24;
    }
  `;

  override async connectedCallback() {
    super.connectedCallback();
    await this.updateStatus();
  }

  private async updateStatus(): Promise<void> {
    this.isLoading = true;
    this.hasError = false;

    const result = await errorHandler.handleAsync(
      async () => {
        const googleResponse = await fetch(API_ENDPOINTS.OAUTH_GOOGLE_CONFIG);
        return googleResponse.ok && (await googleResponse.json()).configured;
      },
      { component: 'OAuthStatus', operation: 'updateStatus' },
      false
    );

    if (result !== undefined) {
      this.googleConfigured = result;
    } else {
      this.hasError = true;
    }

    this.isLoading = false;
  }

  async refresh(): Promise<void> {
    await this.updateStatus();
  }

  private getBadgeStatus() {
    if (this.isLoading) {
      return { text: 'Checking...', className: OAUTH_STATUS.LOADING };
    }
    if (this.hasError) {
      return { text: 'Error', className: OAUTH_STATUS.NOT_CONFIGURED };
    }
    if (this.googleConfigured) {
      return { text: 'Configured', className: OAUTH_STATUS.CONFIGURED };
    }
    return { text: 'Not Configured', className: OAUTH_STATUS.NOT_CONFIGURED };
  }

  override render() {
    const badgeStatus = this.getBadgeStatus();

    return html`
      <div class="oauth-status-section">
        <div class="status-card">
          <div class="status-header">
            <h3>🔐 OAuth Configuration</h3>
            <span class="status-badge ${badgeStatus.className}">${badgeStatus.text}</span>
          </div>
          <p class="status-description">Configure OAuth credentials for external service integrations</p>
          <div class="oauth-providers">
            <div class="provider-status ${this.googleConfigured ? 'configured' : 'not-configured'}">
              <span>🔗</span>
              <span>Google: ${this.googleConfigured ? 'Configured' : 'Not Configured'}</span>
            </div>
          </div>
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'oauth-status': OAuthStatus;
  }
}
