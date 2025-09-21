/**
 * OAuth Status component
 * Shows the current OAuth configuration status
 */

import { ComponentConfig } from '../common/types.js';
import { loadCSS } from '../common/dom-utils.js';
import { logger } from '../common/logger.js';

export class OAuthStatus {
  private container: HTMLElement;
  private config: ComponentConfig;

  constructor(container: HTMLElement, config: ComponentConfig) {
    this.container = container;
    this.config = config;
  }

  async init(): Promise<void> {
    await this.loadStyles();
    this.render();
    await this.updateStatus();
  }

  private async loadStyles(): Promise<void> {
    try {
      await loadCSS('/components/src/integrations/styles/oauth-status.css', 'oauth-status-styles');
    } catch (error) {
      logger.styleError('OAuthStatus', error);
    }
  }

  private render(): void {
    this.container.innerHTML = `
      <div class="oauth-status-section">
        <div class="status-card">
          <div class="status-header">
            <h3>🔐 OAuth Configuration</h3>
            <span id="oauth-status-badge" class="status-badge loading">Checking...</span>
          </div>
          <p class="status-description">Configure OAuth credentials for external service integrations</p>
          <div id="oauth-providers" class="oauth-providers">
            <!-- OAuth providers status will be loaded here -->
          </div>
        </div>
      </div>
    `;
  }

  async updateStatus(): Promise<void> {
    try {
      const googleResponse = await fetch('/api/v1/config/oauth/google');
      const googleConfigured = googleResponse.ok && (await googleResponse.json()).configured;

      const statusBadge = document.getElementById('oauth-status-badge');
      const providersContainer = document.getElementById('oauth-providers');

      if (statusBadge) {
        if (googleConfigured) {
          statusBadge.textContent = 'Configured';
          statusBadge.className = 'status-badge configured';
        } else {
          statusBadge.textContent = 'Not Configured';
          statusBadge.className = 'status-badge not-configured';
        }
      }

      if (providersContainer) {
        providersContainer.innerHTML = `
          <div class="provider-status ${googleConfigured ? 'configured' : 'not-configured'}">
            <span>🔗</span>
            <span>Google: ${googleConfigured ? 'Configured' : 'Not Configured'}</span>
          </div>
        `;
      }
    } catch (error) {
      logger.error('Error updating OAuth status:', error);
      const statusBadge = document.getElementById('oauth-status-badge');
      if (statusBadge) {
        statusBadge.textContent = 'Error';
        statusBadge.className = 'status-badge not-configured';
      }
    }
  }

  async refresh(): Promise<void> {
    await this.updateStatus();
  }

  destroy(): void {
    // Component cleanup if needed
  }
}
