/**
 * OAuthConfigModal Component
 *
 * Role: Modal dialog for configuring OAuth provider credentials
 * Responsibilities:
 * - Provide secure form interface for OAuth client credentials
 * - Display setup instructions for each OAuth provider
 * - Load and save OAuth configuration to backend
 * - Test OAuth configuration validity
 * - Show provider configuration status
 * - Handle form validation and error feedback
 */

import { LitElement, html, css } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { integrationsService } from './integrations-service.js';
import { notifications } from '../common/notification-service.js';
import { errorHandler } from '../common/error-handler.js';
import { buttonStyles, modalStyles, formStyles } from '../common/shared-styles.js';
import { EVENTS } from '../common/constants.js';

@customElement('oauth-config-modal')
export class OAuthConfigModal extends LitElement {
  @state()
  private isVisible = false;

  @state()
  private googleConfigured = false;

  @state()
  private isLoading = false;

  @state()
  private clientId = '';

  @state()
  private clientSecret = '';

  static override styles = [
    buttonStyles,
    modalStyles,
    formStyles,
    css`
      :host {
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        z-index: 1000;
        display: none;
      }

      :host([visible]) {
        display: flex;
      }

      .config-section {
        display: flex;
        flex-direction: column;
        gap: 24px;
      }

      .config-intro {
        margin-bottom: 16px;
      }

      .config-intro h3 {
        margin: 0 0 8px 0;
        font-size: 18px;
        font-weight: 600;
        color: #333;
      }

      .config-intro p {
        margin: 0;
        color: #6c757d;
        line-height: 1.5;
      }

      .provider-config {
        border: 1px solid #e9ecef;
        border-radius: 8px;
        padding: 20px;
        background: #fff;
      }

      .provider-config.disabled {
        background: #f8f9fa;
        opacity: 0.7;
      }

      .provider-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 16px;
      }

      .provider-header h4 {
        margin: 0;
        font-size: 16px;
        font-weight: 600;
        color: #333;
      }

      .config-status {
        padding: 4px 12px;
        border-radius: 16px;
        font-size: 12px;
        font-weight: 600;
        text-transform: uppercase;
      }

      .config-status.configured {
        background: #d4edda;
        color: #155724;
      }

      .config-status.not-configured {
        background: #f8d7da;
        color: #721c24;
      }

      .config-status.coming-soon {
        background: #d1ecf1;
        color: #0c5460;
      }

      .config-instructions {
        margin-bottom: 20px;
      }

      .instructions-details {
        border: 1px solid #e9ecef;
        border-radius: 6px;
        padding: 16px;
        background: #f8f9fa;
      }

      .instructions-details summary {
        cursor: pointer;
        font-weight: 600;
        color: #495057;
      }

      .instructions-details ol {
        margin: 16px 0 0 0;
        padding-left: 20px;
      }

      .instructions-details li {
        margin-bottom: 8px;
        color: #6c757d;
        line-height: 1.4;
      }

      .oauth-form {
        display: flex;
        flex-direction: column;
        gap: 16px;
      }

      .form-row {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 16px;
      }

      .form-actions {
        display: flex;
        gap: 12px;
        justify-content: flex-start;
      }

      .coming-soon-text {
        margin: 0;
        color: #6c757d;
        font-style: italic;
      }

      code {
        background: #f8f9fa;
        padding: 2px 6px;
        border-radius: 4px;
        font-family: 'Monaco', 'Consolas', monospace;
        font-size: 13px;
        color: #e83e8c;
      }

      a {
        color: #007bff;
        text-decoration: none;
      }

      a:hover {
        text-decoration: underline;
      }
    `,
  ];

  override connectedCallback() {
    super.connectedCallback();
    this.addEventListener('keydown', this.handleKeyDown);
  }

  override disconnectedCallback() {
    super.disconnectedCallback();
    this.removeEventListener('keydown', this.handleKeyDown);
  }

  private handleKeyDown = (e: KeyboardEvent) => {
    if (e.key === 'Escape' && this.isVisible) {
      this.hide();
    }
  };

  private handleBackdropClick(e: Event) {
    if (e.target === e.currentTarget) {
      this.hide();
    }
  }

  private handleClientIdInput(e: Event) {
    const target = e.target as HTMLInputElement;
    this.clientId = target.value;
  }

  private handleClientSecretInput(e: Event) {
    const target = e.target as HTMLInputElement;
    this.clientSecret = target.value;
  }

  async show(): Promise<void> {
    this.isVisible = true;
    this.setAttribute('visible', '');
    await this.loadOAuthConfiguration();
  }

  hide(): void {
    this.isVisible = false;
    this.removeAttribute('visible');
    this.resetForm();
  }

  private resetForm(): void {
    this.clientId = '';
    this.clientSecret = '';
  }

  private async loadOAuthConfiguration(): Promise<void> {
    const result = await errorHandler.handleAsync(
      async () => {
        const config = await integrationsService.getOAuthConfig('google');
        this.clientId = config.client_id || '';
        this.googleConfigured = config.configured;
        return true;
      },
      { component: 'OAuthConfigModal', operation: 'loadOAuthConfiguration' },
      false
    );

    if (!result) {
      this.googleConfigured = false;
    }
  }

  private async saveGoogleOAuth(): Promise<void> {
    const clientId = this.clientId.trim();
    const clientSecret = this.clientSecret.trim();

    if (!clientId || !clientSecret) {
      notifications.warning('Please enter both Client ID and Client Secret');
      return;
    }

    this.isLoading = true;

    const success = await errorHandler.handleAsync(
      async () => {
        await integrationsService.updateOAuthConfig('google', {
          client_id: clientId,
          client_secret: clientSecret,
          redirect_url: `${window.location.origin}/oauth/google/callback`,
          scopes: ['https://www.googleapis.com/auth/calendar.readonly'],
          configured: true,
        });

        this.googleConfigured = true;
        this.clientSecret = '';
        notifications.success('Google OAuth configuration saved successfully!');

        this.dispatchEvent(
          new CustomEvent(EVENTS.OAUTH_UPDATED, {
            bubbles: true,
          })
        );
        return true;
      },
      { component: 'OAuthConfigModal', operation: 'saveGoogleOAuth' }
    );

    this.isLoading = false;

    if (!success) {
      notifications.error('Failed to save configuration. Please try again.');
    }
  }

  private async testGoogleOAuth(): Promise<void> {
    const result = await errorHandler.handleAsync(
      async () => {
        const config = await integrationsService.getOAuthConfig('google');
        if (config.configured) {
          notifications.success(
            'Google OAuth appears to be configured correctly. Try connecting Google Calendar from the integrations section to test the full flow.'
          );
        } else {
          notifications.warning('Google OAuth is not properly configured.');
        }
        return true;
      },
      { component: 'OAuthConfigModal', operation: 'testGoogleOAuth' }
    );

    if (!result) {
      notifications.error('Failed to test OAuth configuration.');
    }
  }

  override render() {
    if (!this.isVisible) {
      return html``;
    }

    return html`
      <div class="modal" @click=${this.handleBackdropClick}>
        <div class="modal-content large-modal">
          <div class="modal-header">
            <h2>OAuth Configuration</h2>
            <button class="close-btn" @click=${this.hide} aria-label="Close">&times;</button>
          </div>
          <div class="modal-body">
            <div class="config-section">
              <div class="config-intro">
                <h3>Configure OAuth Providers</h3>
                <p>
                  OAuth credentials allow FamStack to connect to external services on your behalf.
                  Each provider requires you to create an application in their developer console.
                </p>
              </div>

              <!-- Google OAuth Configuration -->
              <div class="provider-config">
                <div class="provider-header">
                  <h4>🔗 Google OAuth</h4>
                  <span
                    class="config-status ${this.googleConfigured ? 'configured' : 'not-configured'}"
                  >
                    ${this.googleConfigured ? 'Configured' : 'Not Configured'}
                  </span>
                </div>

                <div class="config-instructions">
                  <details class="instructions-details">
                    <summary>How to get Google OAuth credentials</summary>
                    <ol>
                      <li>
                        Go to the
                        <a href="https://console.cloud.google.com/" target="_blank"
                          >Google Cloud Console</a
                        >
                      </li>
                      <li>Create a new project or select an existing one</li>
                      <li>Enable the Google Calendar API</li>
                      <li>Go to "Credentials" and create "OAuth 2.0 Client IDs"</li>
                      <li>Set application type to "Web application"</li>
                      <li>
                        Add redirect URI:
                        <code>${window.location.origin}/oauth/google/callback</code>
                      </li>
                      <li>Copy the Client ID and Client Secret below</li>
                    </ol>
                  </details>
                </div>

                <form class="oauth-form" @submit=${(e: Event) => e.preventDefault()}>
                  <div class="form-row">
                    <div class="form-group">
                      <label for="google-client-id">Client ID</label>
                      <input
                        type="text"
                        id="google-client-id"
                        name="client_id"
                        .value=${this.clientId}
                        @input=${this.handleClientIdInput}
                        placeholder="123456789-abc.apps.googleusercontent.com"
                      />
                    </div>
                    <div class="form-group">
                      <label for="google-client-secret">Client Secret</label>
                      <input
                        type="password"
                        id="google-client-secret"
                        name="client_secret"
                        .value=${this.clientSecret}
                        @input=${this.handleClientSecretInput}
                        placeholder="GOCSPX-your-secret-here"
                      />
                    </div>
                  </div>
                  <div class="form-actions">
                    <button
                      type="button"
                      class="btn btn-primary"
                      @click=${this.saveGoogleOAuth}
                      ?disabled=${this.isLoading}
                    >
                      ${this.isLoading ? 'Saving...' : 'Save Google OAuth'}
                    </button>
                    <button
                      type="button"
                      class="btn btn-secondary"
                      @click=${this.testGoogleOAuth}
                      ?disabled=${!this.googleConfigured}
                    >
                      Test Connection
                    </button>
                  </div>
                </form>
              </div>

              <!-- Future providers -->
              <div class="provider-config disabled">
                <div class="provider-header">
                  <h4>🔗 Microsoft OAuth</h4>
                  <span class="config-status coming-soon">Coming Soon</span>
                </div>
                <p class="coming-soon-text">
                  Microsoft OAuth integration will be available in a future update.
                </p>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click=${this.hide}>Close</button>
          </div>
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'oauth-config-modal': OAuthConfigModal;
  }
}
