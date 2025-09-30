/**
 * AddIntegrationModal Component
 *
 * Role: Modal dialog for creating new integrations
 * Responsibilities:
 * - Provide form interface for integration creation
 * - Handle category and provider selection with dynamic updates
 * - Validate form data before submission
 * - Display authentication method information
 * - Dispatch creation events with form data
 * - Manage modal visibility and form reset
 */

import { LitElement, html, css } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import {
  INTEGRATION_CATEGORIES,
  getProvidersByCategory,
  AUTH_DESCRIPTIONS,
} from './integration-types.js';
import { EVENTS } from '../common/constants.js';
import { buttonStyles, modalStyles, formStyles } from '../common/shared-styles.js';

@customElement('add-integration-modal')
export class AddIntegrationModal extends LitElement {
  @state()
  private isVisible = false;

  @state()
  private selectedType = '';

  @state()
  private selectedProvider = '';

  @state()
  private selectedAuthMethod = '';

  @state()
  private providers: Array<{ value: string; label: string; auth: string }> = [];

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

      .auth-info {
        margin-top: 16px;
        padding: 16px;
        background: #f8f9fa;
        border: 1px solid #e9ecef;
        border-radius: 6px;
      }

      .info-box h4 {
        margin: 0 0 8px 0;
        font-size: 14px;
        font-weight: 600;
        color: #495057;
      }

      .info-box p {
        margin: 0;
        font-size: 13px;
        color: #6c757d;
        line-height: 1.4;
      }
    `
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

  private handleTypeChange(e: Event) {
    const target = e.target as HTMLSelectElement;
    this.selectedType = target.value;
    this.selectedProvider = '';
    this.selectedAuthMethod = '';

    if (this.selectedType) {
      this.providers = getProvidersByCategory(this.selectedType);
    } else {
      this.providers = [];
    }
  }

  private handleProviderChange(e: Event) {
    const target = e.target as HTMLSelectElement;
    this.selectedProvider = target.value;

    const selectedOption = target.options[target.selectedIndex];
    this.selectedAuthMethod = selectedOption?.getAttribute('data-auth') || '';
  }

  private async handleCreate() {
    const form = this.shadowRoot?.querySelector('#add-integration-form') as HTMLFormElement;
    if (!form || !form.checkValidity()) {
      form?.reportValidity();
      return;
    }

    const formData = new FormData(form);

    const integrationData = {
      integration_type: formData.get('integration_type'),
      provider: formData.get('provider'),
      auth_method: this.selectedAuthMethod || null,
      display_name: formData.get('display_name'),
      description: formData.get('description') || '',
      settings: {},
    };

    this.dispatchEvent(
      new CustomEvent(EVENTS.CREATE_INTEGRATION, {
        detail: integrationData,
        bubbles: true,
      })
    );

    this.hide();
  }

  show(): void {
    this.isVisible = true;
    this.setAttribute('visible', '');

    this.updateComplete.then(() => {
      const firstInput = this.shadowRoot?.querySelector('select, input') as HTMLElement;
      if (firstInput) {
        firstInput.focus();
      }
    });
  }

  hide(): void {
    this.isVisible = false;
    this.removeAttribute('visible');
    this.resetForm();
  }

  private resetForm(): void {
    this.selectedType = '';
    this.selectedProvider = '';
    this.selectedAuthMethod = '';
    this.providers = [];

    this.updateComplete.then(() => {
      const form = this.shadowRoot?.querySelector('#add-integration-form') as HTMLFormElement;
      form?.reset();
    });
  }

  override render() {
    if (!this.isVisible) {
      return html``;
    }

    return html`
      <div class="modal" @click=${this.handleBackdropClick}>
        <div class="modal-content">
          <div class="modal-header">
            <h2>Add New Integration</h2>
            <button class="close-btn" @click=${this.hide} aria-label="Close">&times;</button>
          </div>
          <div class="modal-body">
            <form id="add-integration-form">
              <div class="form-group">
                <label for="integration-type">Category</label>
                <select
                  id="integration-type"
                  name="integration_type"
                  required
                  .value=${this.selectedType}
                  @change=${this.handleTypeChange}
                >
                  <option value="">Select a category</option>
                  ${INTEGRATION_CATEGORIES.map(
                    cat => html`<option value="${cat.key}">${cat.label}</option>`
                  )}
                </select>
              </div>

              <div class="form-group">
                <label for="provider">Provider</label>
                <select
                  id="provider"
                  name="provider"
                  required
                  .value=${this.selectedProvider}
                  @change=${this.handleProviderChange}
                >
                  <option value="">Select a provider</option>
                  ${this.providers.map(
                    provider => html`
                      <option value="${provider.value}" data-auth="${provider.auth}">
                        ${provider.label}
                      </option>
                    `
                  )}
                </select>
              </div>

              <div class="form-group">
                <label for="display-name">Display Name</label>
                <input
                  type="text"
                  id="display-name"
                  name="display_name"
                  required
                  placeholder="e.g., John's Google Calendar"
                />
              </div>

              <div class="form-group">
                <label for="description">Description (optional)</label>
                <textarea
                  id="description"
                  name="description"
                  rows="3"
                  placeholder="Brief description of this integration"
                ></textarea>
              </div>

              ${this.selectedAuthMethod
                ? html`
                    <div class="auth-info">
                      <div class="info-box">
                        <h4>Authentication Method: ${this.selectedAuthMethod.toUpperCase()}</h4>
                        <p>
                          ${AUTH_DESCRIPTIONS[this.selectedAuthMethod] ||
                          'Authentication method information not available.'}
                        </p>
                      </div>
                    </div>
                  `
                : ''}
            </form>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click=${this.hide}>Cancel</button>
            <button class="btn btn-primary" @click=${this.handleCreate}>Add Integration</button>
          </div>
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'add-integration-modal': AddIntegrationModal;
  }
}
