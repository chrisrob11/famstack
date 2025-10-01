/**
 * Integrations page component for SPA
 */

import { BasePage } from './base-page.js';
import { ComponentConfig } from '../common/types.js';
import '../integrations/oauth-status.js';
import '../integrations/category-tabs.js';
import '../integrations/integration-grid.js';
import '../integrations/add-integration-modal.js';
import '../integrations/oauth-config-modal.js';
import '../integrations/integration-details-modal.js';
import { logger } from '../common/logger.js';

export class IntegrationsPage extends BasePage {
  constructor(container: HTMLElement, config: ComponentConfig) {
    super(container, config, 'integrations');
  }

  async init(): Promise<void> {
    this.render();
    this.setupEventListeners();
    await this.initializeComponents();
  }

  private render(): void {
    this.container.innerHTML = `
      <div class="integrations-page">
        <div class="page-header">
          <div class="header-content">
            <h1>Integrations</h1>
            <p>Connect external services to enhance your family management experience</p>
          </div>
          <div class="header-actions">
            <button class="btn btn-secondary" id="oauth-settings-btn">
              ⚙️ OAuth Settings
            </button>
            <button class="btn btn-primary" id="add-integration-btn">
              + Add Integration
            </button>
          </div>
        </div>

        <!-- OAuth Status -->
        <oauth-status></oauth-status>

        <!-- Category Tabs -->
        <category-tabs active-category="all"></category-tabs>

        <!-- Integration Grid -->
        <integration-grid current-category="all"></integration-grid>

        <!-- Modals -->
        <add-integration-modal></add-integration-modal>
        <oauth-config-modal></oauth-config-modal>
        <integration-details-modal></integration-details-modal>
      </div>
    `;

    this.addStyles();
  }

  private addStyles(): void {
    if (document.getElementById('integrations-page-styles')) return;

    const styles = `
      <style id="integrations-page-styles">
        .integrations-page {
          padding: 2rem;
          max-width: 1200px;
          margin: 0 auto;
        }

        .page-header {
          display: flex;
          justify-content: space-between;
          align-items: flex-start;
          margin-bottom: 2rem;
          flex-wrap: wrap;
          gap: 1rem;
        }

        .header-content h1 {
          font-size: 2rem;
          font-weight: 700;
          color: #374151;
          margin: 0 0 0.5rem 0;
        }

        .header-content p {
          color: #6b7280;
          font-size: 1rem;
          margin: 0;
        }

        .header-actions {
          display: flex;
          gap: 0.75rem;
          flex-wrap: wrap;
        }

        .btn {
          padding: 0.75rem 1.5rem;
          border-radius: 0.5rem;
          font-size: 0.875rem;
          font-weight: 500;
          border: none;
          cursor: pointer;
          transition: all 0.2s;
        }

        .btn-primary {
          background: #3b82f6;
          color: white;
        }

        .btn-primary:hover:not(:disabled) {
          background: #2563eb;
        }

        .btn-secondary {
          background: #f3f4f6;
          color: #374151;
          border: 1px solid #d1d5db;
        }

        .btn-secondary:hover:not(:disabled) {
          background: #e5e7eb;
        }

        @media (max-width: 768px) {
          .page-header {
            flex-direction: column;
            align-items: stretch;
          }

          .header-actions {
            width: 100%;
            justify-content: flex-start;
          }
        }
      </style>
    `;

    document.head.insertAdjacentHTML('beforeend', styles);
  }

  private setupEventListeners(): void {
    // Header buttons
    const oauthSettingsBtn = document.getElementById('oauth-settings-btn');
    const addIntegrationBtn = document.getElementById('add-integration-btn');

    if (oauthSettingsBtn) {
      oauthSettingsBtn.addEventListener('click', () => {
        const modal = this.container.querySelector('oauth-config-modal') as any;
        if (modal) {
          modal.show();
        }
      });
    }

    if (addIntegrationBtn) {
      addIntegrationBtn.addEventListener('click', () => {
        const modal = this.container.querySelector('add-integration-modal') as any;
        if (modal) {
          modal.show();
        }
      });
    }

    // Component event listeners
    this.container.addEventListener('category-changed', (e: any) => {
      const integrationGrid = this.container.querySelector('integration-grid') as any;
      if (integrationGrid) {
        integrationGrid.setAttribute('current-category', e.detail.category);
      }
    });

    this.container.addEventListener('create-integration', async (e: any) => {
      try {
        const integrationGrid = this.container.querySelector('integration-grid') as any;
        if (integrationGrid) {
          await integrationGrid.addIntegration(e.detail);
        }
      } catch (error) {
        logger.error('Failed to create integration:', error);
        alert(
          `Failed to create integration: ${error instanceof Error ? error.message : 'Unknown error'}`
        );
      }
    });

    this.container.addEventListener('delete-integration', async (e: any) => {
      if (confirm('Are you sure you want to delete this integration?')) {
        try {
          const integrationGrid = this.container.querySelector('integration-grid') as any;
          if (integrationGrid) {
            await integrationGrid.deleteIntegration(e.detail.id);
          }
        } catch (error) {
          logger.error('Failed to delete integration:', error);
          alert(
            `Failed to delete integration: ${error instanceof Error ? error.message : 'Unknown error'}`
          );
        }
      }
    });

    this.container.addEventListener('configure-integration', (_e: any) => {
      const modal = this.container.querySelector('oauth-config-modal') as any;
      if (modal) {
        modal.show();
      }
    });

    this.container.addEventListener('oauth-updated', () => {
      const oauthStatus = this.container.querySelector('oauth-status') as any;
      if (oauthStatus) {
        oauthStatus.refresh();
      }
    });

    // Handle clicking on integration cards to show details
    this.container.addEventListener('click', (e: any) => {
      const card = e.target.closest('integration-card');
      if (card && card.integration) {
        const detailsModal = this.container.querySelector('integration-details-modal') as any;
        if (detailsModal) {
          detailsModal.show(card.integration.id);
        }
      }
    });

    // Handle integration connected event to refresh data
    this.container.addEventListener('integration-connected', async () => {
      await this.refresh();
    });

    // Handle integration synced event
    this.container.addEventListener('integration-synced', async () => {
      await this.refresh();
    });
  }

  private async initializeComponents(): Promise<void> {
    // Components are now initialized automatically as custom elements
    // No explicit initialization needed
  }

  async refresh(): Promise<void> {
    const oauthStatus = this.container.querySelector('oauth-status') as any;
    if (oauthStatus && oauthStatus.refresh) {
      await oauthStatus.refresh();
    }
    const integrationGrid = this.container.querySelector('integration-grid') as any;
    if (integrationGrid && integrationGrid.refresh) {
      await integrationGrid.refresh();
    }
  }

  override destroy(): void {
    // Remove styles
    const styles = document.getElementById('integrations-page-styles');
    if (styles) {
      styles.remove();
    }

    super.destroy();
  }
}

export default IntegrationsPage;
