/**
 * Add Integration Modal component
 * Modal for creating new integrations
 */

import { ComponentConfig } from '../common/types.js';
import {
  INTEGRATION_CATEGORIES,
  getProvidersByCategory,
  AUTH_DESCRIPTIONS,
} from './integration-types.js';
import { loadCSS } from '../common/dom-utils.js';
import { logger } from '../common/logger.js';

export class AddIntegrationModal {
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
        loadCSS(
          '/components/src/integrations/styles/add-integration.css',
          'add-integration-styles'
        ),
      ]);
    } catch (error) {
      logger.styleError('AddIntegrationModal', error);
    }
  }

  private render(): void {
    const modalHtml = `
      <div id="add-integration-modal" class="modal" style="display: none;">
        <div class="modal-content">
          <div class="modal-header">
            <h2>Add New Integration</h2>
            <button class="close-btn" data-action="close">&times;</button>
          </div>
          <div class="modal-body">
            <form id="add-integration-form">
              <div class="form-group">
                <label for="integration-type">Category</label>
                <select id="integration-type" name="integration_type" required>
                  <option value="">Select a category</option>
                  ${INTEGRATION_CATEGORIES.map(
                    cat => `<option value="${cat.key}">${cat.label}</option>`
                  ).join('')}
                </select>
              </div>

              <div class="form-group">
                <label for="provider">Provider</label>
                <select id="provider" name="provider" required>
                  <option value="">Select a provider</option>
                </select>
              </div>

              <div class="form-group">
                <label for="display-name">Display Name</label>
                <input type="text" id="display-name" name="display_name" required
                       placeholder="e.g., John's Google Calendar">
              </div>

              <div class="form-group">
                <label for="description">Description (optional)</label>
                <textarea id="description" name="description" rows="3"
                          placeholder="Brief description of this integration"></textarea>
              </div>

              <div id="auth-method-info" class="auth-info" style="display: none;">
                <div class="info-box">
                  <h4>Authentication Method: <span id="auth-method-name"></span></h4>
                  <p id="auth-method-description"></p>
                </div>
              </div>
            </form>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" data-action="close">Cancel</button>
            <button class="btn btn-primary" id="create-integration-btn">Add Integration</button>
          </div>
        </div>
      </div>
    `;

    this.container.insertAdjacentHTML('beforeend', modalHtml);
    this.modal = document.getElementById('add-integration-modal') as HTMLElement;
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

    // Form handlers
    const integrationTypeSelect = document.getElementById('integration-type');
    const providerSelect = document.getElementById('provider');
    const createBtn = document.getElementById('create-integration-btn');

    if (integrationTypeSelect) {
      integrationTypeSelect.addEventListener('change', () => this.updateProviders());
    }

    if (providerSelect) {
      providerSelect.addEventListener('change', () => this.updateAuthMethod());
    }

    if (createBtn) {
      createBtn.addEventListener('click', () => this.handleCreate());
    }

    // Escape key to close
    document.addEventListener('keydown', e => {
      if (e.key === 'Escape' && this.isVisible()) {
        this.hide();
      }
    });
  }

  private updateProviders(): void {
    const typeSelect = document.getElementById('integration-type') as HTMLSelectElement;
    const providerSelect = document.getElementById('provider') as HTMLSelectElement;
    const selectedType = typeSelect.value;

    // Clear existing options
    providerSelect.innerHTML = '<option value="">Select a provider</option>';

    if (selectedType) {
      const providers = getProvidersByCategory(selectedType);
      providers.forEach(provider => {
        const option = document.createElement('option');
        option.value = provider.value;
        option.textContent = provider.label;
        option.setAttribute('data-auth', provider.auth);
        providerSelect.appendChild(option);
      });
    }

    // Clear auth method info
    const authInfo = document.getElementById('auth-method-info');
    if (authInfo) {
      authInfo.style.display = 'none';
    }
  }

  private updateAuthMethod(): void {
    const providerSelect = document.getElementById('provider') as HTMLSelectElement;
    const selectedOption = providerSelect.options[providerSelect.selectedIndex];

    if (selectedOption && selectedOption.getAttribute('data-auth')) {
      const authMethod = selectedOption.getAttribute('data-auth')!;
      const authMethodName = document.getElementById('auth-method-name');
      const authMethodDescription = document.getElementById('auth-method-description');
      const authInfo = document.getElementById('auth-method-info');

      if (authMethodName && authMethodDescription && authInfo) {
        authMethodName.textContent = authMethod.toUpperCase();
        authMethodDescription.textContent =
          AUTH_DESCRIPTIONS[authMethod] || 'Authentication method information not available.';
        authInfo.style.display = 'block';
      }
    } else {
      const authInfo = document.getElementById('auth-method-info');
      if (authInfo) {
        authInfo.style.display = 'none';
      }
    }
  }

  private async handleCreate(): Promise<void> {
    const form = document.getElementById('add-integration-form') as HTMLFormElement;
    if (!form.checkValidity()) {
      form.reportValidity();
      return;
    }

    const formData = new FormData(form);
    const providerSelect = document.getElementById('provider') as HTMLSelectElement;
    const selectedOption = providerSelect.options[providerSelect.selectedIndex];

    const integrationData = {
      integration_type: formData.get('integration_type'),
      provider: formData.get('provider'),
      auth_method: selectedOption?.getAttribute('data-auth') || null,
      display_name: formData.get('display_name'),
      description: formData.get('description') || '',
      settings: {},
    };

    // Dispatch event to parent component
    this.container.dispatchEvent(
      new CustomEvent('create-integration', {
        detail: integrationData,
        bubbles: true,
      })
    );
  }

  show(): void {
    if (this.modal) {
      this.modal.style.display = 'flex';

      // Focus on first input
      const firstInput = this.modal.querySelector('select, input') as HTMLElement;
      if (firstInput) {
        firstInput.focus();
      }
    }
  }

  hide(): void {
    if (this.modal) {
      this.modal.style.display = 'none';
      this.resetForm();
    }
  }

  private resetForm(): void {
    const form = document.getElementById('add-integration-form') as HTMLFormElement;
    if (form) {
      form.reset();
    }

    const authInfo = document.getElementById('auth-method-info');
    if (authInfo) {
      authInfo.style.display = 'none';
    }

    const providerSelect = document.getElementById('provider') as HTMLSelectElement;
    if (providerSelect) {
      providerSelect.innerHTML = '<option value="">Select a provider</option>';
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
