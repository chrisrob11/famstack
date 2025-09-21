/**
 * Integrations page component for SPA
 */

import { BasePage } from './base-page.js';
import { ComponentConfig } from '../common/types.js';
import { OAuthStatus } from '../integrations/oauth-status.js';
import { CategoryTabs } from '../integrations/category-tabs.js';
import { IntegrationGrid } from '../integrations/integration-grid.js';
import { AddIntegrationModal } from '../integrations/add-integration-modal.js';
import { OAuthConfigModal } from '../integrations/oauth-config-modal.js';
import { logger } from '../common/logger.js';

export class IntegrationsPage extends BasePage {
  private oauthStatus?: OAuthStatus;
  private categoryTabs?: CategoryTabs;
  private integrationGrid?: IntegrationGrid;
  private addIntegrationModal?: AddIntegrationModal;
  private oauthConfigModal?: OAuthConfigModal;

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

        <!-- OAuth Status will be initialized here -->
        <div id="oauth-status-container"></div>

        <!-- Category Tabs will be initialized here -->
        <div id="category-tabs-container"></div>

        <!-- Integration Grid will be initialized here -->
        <div id="integration-grid-container"></div>
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
        if (this.oauthConfigModal) {
          this.oauthConfigModal.show();
        }
      });
    }

    if (addIntegrationBtn) {
      addIntegrationBtn.addEventListener('click', () => {
        if (this.addIntegrationModal) {
          this.addIntegrationModal.show();
        }
      });
    }

    // Component event listeners
    this.container.addEventListener('category-changed', (e: any) => {
      if (this.integrationGrid) {
        this.integrationGrid.setCategory(e.detail.category);
      }
    });

    this.container.addEventListener('create-integration', async (e: any) => {
      try {
        if (this.integrationGrid) {
          await this.integrationGrid.addIntegration(e.detail);
          if (this.addIntegrationModal) {
            this.addIntegrationModal.hide();
          }
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
          if (this.integrationGrid) {
            await this.integrationGrid.deleteIntegration(e.detail.id);
          }
        } catch (error) {
          logger.error('Failed to delete integration:', error);
          alert(
            `Failed to delete integration: ${error instanceof Error ? error.message : 'Unknown error'}`
          );
        }
      }
    });

    this.container.addEventListener('oauth-updated', () => {
      if (this.oauthStatus) {
        this.oauthStatus.refresh();
      }
    });
  }

  private async initializeComponents(): Promise<void> {
    // Initialize OAuth Status
    const oauthStatusContainer = document.getElementById('oauth-status-container');
    if (oauthStatusContainer) {
      this.oauthStatus = new OAuthStatus(oauthStatusContainer, this.config);
      await this.oauthStatus.init();
    }

    // Initialize Category Tabs
    const categoryTabsContainer = document.getElementById('category-tabs-container');
    if (categoryTabsContainer) {
      this.categoryTabs = new CategoryTabs(categoryTabsContainer, this.config);
      this.categoryTabs.init();
    }

    // Initialize Integration Grid
    const integrationGridContainer = document.getElementById('integration-grid-container');
    if (integrationGridContainer) {
      this.integrationGrid = new IntegrationGrid(integrationGridContainer, this.config);
      await this.integrationGrid.init();
    }

    // Initialize Modals
    this.addIntegrationModal = new AddIntegrationModal(this.container, this.config);
    this.addIntegrationModal.init();

    this.oauthConfigModal = new OAuthConfigModal(this.container, this.config);
    this.oauthConfigModal.init();
  }

  async refresh(): Promise<void> {
    if (this.oauthStatus) {
      await this.oauthStatus.refresh();
    }
    if (this.integrationGrid) {
      await this.integrationGrid.refresh();
    }
  }

  override destroy(): void {
    if (this.oauthStatus) {
      this.oauthStatus.destroy();
    }
    if (this.categoryTabs) {
      this.categoryTabs.destroy();
    }
    if (this.integrationGrid) {
      this.integrationGrid.destroy();
    }
    if (this.addIntegrationModal) {
      this.addIntegrationModal.destroy();
    }
    if (this.oauthConfigModal) {
      this.oauthConfigModal.destroy();
    }

    // Remove styles
    const styles = document.getElementById('integrations-page-styles');
    if (styles) {
      styles.remove();
    }

    super.destroy();
  }
}

export default IntegrationsPage;
