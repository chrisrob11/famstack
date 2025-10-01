/**
 * FamilyInfo Lit Component
 *
 * Component for creating and viewing family information
 */

import { LitElement, html, css } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { buttonStyles, formStyles } from '../common/shared-styles.js';
import { errorHandler } from '../common/error-handler.js';
import { familyContext, Family } from './family-context.js';
import { showToast } from '../common/toast-notification.js';

@customElement('family-info')
export class FamilyInfo extends LitElement {
  @state()
  private currentFamily: Family | null = null;

  @state()
  private isLoadingFamily = true;

  @state()
  private isSubmitting = false;

  @state()
  private isEditing = false;

  @state()
  private errorMessage = '';

  static override styles = [
    buttonStyles,
    formStyles,
    css`
      :host {
        display: block;
      }

      .family-info-section {
        background: white;
        border: 1px solid #e5e7eb;
        border-radius: 8px;
        padding: 24px;
        margin-bottom: 24px;
      }

      h3 {
        margin: 0 0 16px 0;
        font-size: 20px;
        font-weight: 600;
        color: #374151;
      }

      .success-message {
        padding: 12px;
        background: #d1fae5;
        color: #065f46;
        border-radius: 4px;
        margin-bottom: 16px;
        font-size: 14px;
      }

      .error-message {
        padding: 12px;
        background: #fecaca;
        color: #991b1b;
        border-radius: 4px;
        margin-bottom: 16px;
        font-size: 14px;
      }

      .family-form {
        max-width: 500px;
      }

      .current-family {
        background: #f9fafb;
        border: 1px solid #e5e7eb;
        border-radius: 6px;
        padding: 16px;
        margin-bottom: 20px;
      }

      .family-details {
        display: flex;
        justify-content: space-between;
        align-items: start;
        margin-bottom: 12px;
      }

      .family-name {
        font-size: 18px;
        font-weight: 600;
        color: #111827;
        margin: 0 0 4px 0;
      }

      .family-meta {
        font-size: 13px;
        color: #6b7280;
        margin: 0;
      }

      .family-actions {
        display: flex;
        gap: 8px;
      }

      .loading-state {
        text-align: center;
        padding: 20px;
        color: #6b7280;
      }

      .no-family {
        text-align: center;
        padding: 20px;
        color: #6b7280;
      }

      .no-family p {
        margin: 0 0 16px 0;
      }
    `
  ];

  override async connectedCallback() {
    super.connectedCallback();
    await this.loadCurrentFamily();
  }

  private async loadCurrentFamily() {
    this.isLoadingFamily = true;

    const family = await errorHandler.handleAsync(
      async () => {
        return await familyContext.getFamily();
      },
      { component: 'FamilyInfo', operation: 'loadFamily' }
    );

    this.currentFamily = family || null;
    this.isLoadingFamily = false;
  }

  private async handleSubmit(e: Event) {
    e.preventDefault();
    if (this.isSubmitting) return;

    const form = e.target as HTMLFormElement;
    const formData = new FormData(form);
    const name = (formData.get('name') as string).trim();

    if (!name) {
      this.errorMessage = 'Family name is required';
      return;
    }

    this.isSubmitting = true;
    this.errorMessage = '';

    if (this.isEditing && this.currentFamily) {
      // Update existing family
      const result = await errorHandler.handleAsync(
        async () => {
          const response = await fetch(`/api/v1/families/${this.currentFamily!.id}`, {
            method: 'PATCH',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({ name }),
          });

          if (!response.ok) {
            throw new Error('Failed to update family');
          }

          return await response.json();
        },
        { component: 'FamilyInfo', operation: 'updateFamily' }
      );

      this.isSubmitting = false;

      if (result) {
        this.currentFamily = result;
        familyContext.updateFamily(result);
        showToast(`Family name updated to "${result.name}"`, 'success');
        this.isEditing = false;
        form.reset();

        this.dispatchEvent(new CustomEvent('family-updated', {
          detail: { family: result },
          bubbles: true
        }));
      }
    } else {
      // Create new family
      const result = await errorHandler.handleAsync(
        async () => {
          const response = await fetch('/api/v1/families', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({ name }),
          });

          if (!response.ok) {
            throw new Error('Failed to create family');
          }

          return await response.json();
        },
        { component: 'FamilyInfo', operation: 'createFamily' }
      );

      this.isSubmitting = false;

      if (result) {
        this.currentFamily = result;
        familyContext.updateFamily(result);
        showToast(`Family "${result.name}" created successfully!`, 'success');
        form.reset();

        this.dispatchEvent(new CustomEvent('family-created', {
          detail: { family: result },
          bubbles: true
        }));
      }
    }
  }

  private handleEdit() {
    this.isEditing = true;
    this.errorMessage = '';
  }

  private handleCancelEdit() {
    this.isEditing = false;
    this.errorMessage = '';
  }

  private renderCurrentFamily() {
    if (this.isLoadingFamily) {
      return html`
        <div class="loading-state">
          <p>Loading family information...</p>
        </div>
      `;
    }

    if (!this.currentFamily) {
      return html`
        <div class="no-family">
          <p>You don't have a family yet. Create one below to get started!</p>
        </div>
      `;
    }

    const createdDate = new Date(this.currentFamily.created_at).toLocaleDateString();

    return html`
      <div class="current-family">
        <div class="family-details">
          <div>
            <h4 class="family-name">${this.currentFamily.name}</h4>
            <p class="family-meta">Created on ${createdDate}</p>
          </div>
          <div class="family-actions">
            <button
              class="btn btn-sm btn-secondary"
              @click=${this.handleEdit}
              ?disabled=${this.isEditing}
            >
              Edit
            </button>
          </div>
        </div>
      </div>
    `;
  }

  private renderForm() {
    // Only show form if editing or no family exists
    if (this.currentFamily && !this.isEditing) {
      return html``;
    }

    const isEdit = this.isEditing && this.currentFamily;
    const submitText = this.isSubmitting
      ? (isEdit ? 'Updating...' : 'Creating...')
      : (isEdit ? 'Update Family Name' : 'Create New Family');

    return html`
      ${this.errorMessage
        ? html`<div class="error-message">${this.errorMessage}</div>`
        : ''}

      <form class="family-form" @submit=${this.handleSubmit}>
        <div class="form-group">
          <label for="family-name">Family Name</label>
          <input
            type="text"
            id="family-name"
            name="name"
            .value=${isEdit ? this.currentFamily!.name : ''}
            placeholder="Enter your family name"
            required
          />
        </div>
        <div style="display: flex; gap: 12px;">
          <button type="submit" class="btn btn-primary" ?disabled=${this.isSubmitting}>
            ${submitText}
          </button>
          ${isEdit
            ? html`
                <button
                  type="button"
                  class="btn btn-secondary"
                  @click=${this.handleCancelEdit}
                  ?disabled=${this.isSubmitting}
                >
                  Cancel
                </button>
              `
            : ''}
        </div>
      </form>
    `;
  }

  override render() {
    return html`
      <section class="family-info-section">
        <h3>Current Family</h3>
        ${this.renderCurrentFamily()}
        ${this.renderForm()}
      </section>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'family-info': FamilyInfo;
  }
}
