/**
 * OAuth Configuration Modal component
 * Modal for configuring OAuth settings
 */

import { ComponentConfig } from '../common/types.js';
import { integrationsService } from './integrations-service.js';
import { loadCSS } from '../common/dom-utils.js';
import { notifications } from '../common/notification-service.js';
import { logger } from '../common/logger.js';

export class OAuthConfigModal {
  private container: HTMLElement;
  private config: ComponentConfig;
  private modal?: HTMLElement;

  constructor(container: HTMLElement, config: ComponentConfig) {
    this.container = container;
    this.config = config;
  }

  async init(): Promise<void> {
    await this.loadStyles();
    this.render();
    this.setupEventListeners();
  }

  private async loadStyles(): Promise<void> {
    try {
      await Promise.all([
        loadCSS('/components/src/integrations/styles/modal.css', 'modal-styles'),
        loadCSS('/components/src/integrations/styles/oauth-config.css', 'oauth-config-styles'),
      ]);
    } catch (error) {
      logger.styleError('OAuthConfigModal', error);
    }
  }

  private render(): void {
    const modalHtml = `
      <div id="oauth-config-modal" class="modal" style="display: none;">
        <div class="modal-content large-modal">
          <div class="modal-header">
            <h2>OAuth Configuration</h2>
            <button class="close-btn" data-action="close">&times;</button>
          </div>
          <div class="modal-body">
            <div class="config-section">
              <div class="config-intro">
                <h3>Configure OAuth Providers</h3>
                <p>OAuth credentials allow FamStack to connect to external services on your behalf.
                Each provider requires you to create an application in their developer console.</p>
              </div>

              <!-- Google OAuth Configuration -->
              <div class="provider-config">
                <div class="provider-header">
                  <h4>🔗 Google OAuth</h4>
                  <span id="google-config-status" class="config-status not-configured">Not Configured</span>
                </div>

                <div class="config-instructions">
                  <details class="instructions-details">
                    <summary>How to get Google OAuth credentials</summary>
                    <ol>
                      <li>Go to the <a href="https://console.cloud.google.com/" target="_blank">Google Cloud Console</a></li>
                      <li>Create a new project or select an existing one</li>
                      <li>Enable the Google Calendar API</li>
                      <li>Go to "Credentials" and create "OAuth 2.0 Client IDs"</li>
                      <li>Set application type to "Web application"</li>
                      <li>Add redirect URI: <code id="google-redirect-uri">${window.location.origin}/oauth/google/callback</code></li>
                      <li>Copy the Client ID and Client Secret below</li>
                    </ol>
                  </details>
                </div>

                <form id="google-oauth-form" class="oauth-form">
                  <div class="form-row">
                    <div class="form-group">
                      <label for="google-client-id">Client ID</label>
                      <input type="text" id="google-client-id" name="client_id"
                             placeholder="123456789-abc.apps.googleusercontent.com">
                    </div>
                    <div class="form-group">
                      <label for="google-client-secret">Client Secret</label>
                      <input type="password" id="google-client-secret" name="client_secret"
                             placeholder="GOCSPX-your-secret-here">
                    </div>
                  </div>
                  <div class="form-actions">
                    <button type="button" class="btn btn-primary" id="save-google-oauth-btn">
                      Save Google OAuth
                    </button>
                    <button type="button" class="btn btn-secondary" id="test-google-btn" disabled>
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
                <p class="coming-soon-text">Microsoft OAuth integration will be available in a future update.</p>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" data-action="close">Close</button>
          </div>
        </div>
      </div>
    `;

    this.container.insertAdjacentHTML('beforeend', modalHtml);
    this.modal = document.getElementById('oauth-config-modal') as HTMLElement;
  }

  private setupEventListeners(): void {
    if (!this.modal) return;

    // Close modal handlers
    this.modal.addEventListener('click', e => {
      const target = e.target as HTMLElement;

      if (target.getAttribute('data-action') === 'close' || target === this.modal) {
        this.hide();
      }
    });

    // OAuth configuration handlers
    const saveGoogleOAuthBtn = document.getElementById('save-google-oauth-btn');
    const testGoogleBtn = document.getElementById('test-google-btn');

    if (saveGoogleOAuthBtn) {
      saveGoogleOAuthBtn.addEventListener('click', () => this.saveGoogleOAuth());
    }

    if (testGoogleBtn) {
      testGoogleBtn.addEventListener('click', () => this.testGoogleOAuth());
    }

    // Escape key to close
    document.addEventListener('keydown', e => {
      if (e.key === 'Escape' && this.isVisible()) {
        this.hide();
      }
    });
  }

  async show(): Promise<void> {
    if (this.modal) {
      this.modal.style.display = 'flex';
      await this.loadOAuthConfiguration();
    }
  }

  hide(): void {
    if (this.modal) {
      this.modal.style.display = 'none';
    }
  }

  private async loadOAuthConfiguration(): Promise<void> {
    try {
      // Load Google OAuth config
      const config = await integrationsService.getOAuthConfig('google');
      const clientIdInput = document.getElementById('google-client-id') as HTMLInputElement;
      const testBtn = document.getElementById('test-google-btn') as HTMLButtonElement;

      if (clientIdInput) {
        clientIdInput.value = config.client_id || '';
      }

      this.updateGoogleOAuthStatus(config.configured);

      if (testBtn) {
        testBtn.disabled = !config.configured;
      }
    } catch (error) {
      logger.error('Error loading OAuth configuration:', error);
      this.updateGoogleOAuthStatus(false);
    }
  }

  private updateGoogleOAuthStatus(isConfigured: boolean): void {
    const statusElement = document.getElementById('google-config-status');
    if (statusElement) {
      if (isConfigured) {
        statusElement.textContent = 'Configured';
        statusElement.className = 'config-status configured';
      } else {
        statusElement.textContent = 'Not Configured';
        statusElement.className = 'config-status not-configured';
      }
    }
  }

  private async saveGoogleOAuth(): Promise<void> {
    const clientIdInput = document.getElementById('google-client-id') as HTMLInputElement;
    const clientSecretInput = document.getElementById('google-client-secret') as HTMLInputElement;

    const clientId = clientIdInput?.value.trim();
    const clientSecret = clientSecretInput?.value.trim();

    if (!clientId || !clientSecret) {
      notifications.warning('Please enter both Client ID and Client Secret');
      return;
    }

    try {
      await integrationsService.updateOAuthConfig('google', {
        client_id: clientId,
        client_secret: clientSecret,
        redirect_url: `${window.location.origin}/oauth/google/callback`,
        scopes: ['https://www.googleapis.com/auth/calendar.readonly'],
        configured: true,
      });

      this.updateGoogleOAuthStatus(true);
      const testBtn = document.getElementById('test-google-btn') as HTMLButtonElement;
      if (testBtn) {
        testBtn.disabled = false;
      }
      if (clientSecretInput) {
        clientSecretInput.value = ''; // Clear for security
      }
      notifications.success('Google OAuth configuration saved successfully!');

      // Dispatch event to update OAuth status
      this.container.dispatchEvent(
        new CustomEvent('oauth-updated', {
          bubbles: true,
        })
      );
    } catch (error) {
      logger.error('Error saving OAuth configuration:', error);
      const errorMessage =
        error instanceof Error ? error.message : 'Failed to save configuration. Please try again.';
      notifications.error(errorMessage);
    }
  }

  private async testGoogleOAuth(): Promise<void> {
    try {
      const config = await integrationsService.getOAuthConfig('google');
      if (config.configured) {
        notifications.success(
          'Google OAuth appears to be configured correctly. Try connecting Google Calendar from the integrations section to test the full flow.'
        );
      } else {
        notifications.warning('Google OAuth is not properly configured.');
      }
    } catch (error) {
      logger.error('Error testing OAuth:', error);
      notifications.error('Failed to test OAuth configuration.');
    }
  }

  private isVisible(): boolean {
    return this.modal ? this.modal.style.display === 'flex' : false;
  }

  destroy(): void {
    if (this.modal) {
      this.modal.remove();
    }
  }
}
