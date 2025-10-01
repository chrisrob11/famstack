/**
 * Confirmation Modal Component
 *
 * Reusable modal for confirming destructive actions
 */

import { LitElement, html, css } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { modalStyles, buttonStyles } from './shared-styles.js';

@customElement('confirmation-modal')
export class ConfirmationModal extends LitElement {
  @property({ type: Boolean })
  open = false;

  @property({ type: String })
  override title = 'Confirm Action';

  @property({ type: String })
  message = 'Are you sure you want to proceed?';

  @property({ type: String })
  confirmText = 'Confirm';

  @property({ type: String })
  cancelText = 'Cancel';

  @property({ type: String })
  variant: 'danger' | 'warning' | 'info' = 'danger';

  @state()
  private resolvePromise: ((value: boolean) => void) | undefined = undefined;

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

      .modal-content {
        max-width: 450px;
      }

      .message {
        margin: 0;
        color: #374151;
        font-size: 14px;
        line-height: 1.6;
      }

      .modal-footer {
        display: flex;
        justify-content: flex-end;
        gap: 12px;
        padding: 20px;
        border-top: 1px solid #e1e5e9;
        background: #f8f9fa;
      }

      .btn-danger {
        background: #dc3545;
        border-color: #dc3545;
        color: white;
      }

      .btn-danger:hover:not(:disabled) {
        background: #c82333;
        border-color: #bd2130;
      }

      .btn-warning {
        background: #ffc107;
        border-color: #ffc107;
        color: #000;
      }

      .btn-warning:hover:not(:disabled) {
        background: #e0a800;
        border-color: #d39e00;
      }

      .btn-info {
        background: #17a2b8;
        border-color: #17a2b8;
        color: white;
      }

      .btn-info:hover:not(:disabled) {
        background: #138496;
        border-color: #117a8b;
      }
    `
  ];

  async confirm(options: {
    title?: string;
    message: string;
    confirmText?: string;
    cancelText?: string;
    variant?: 'danger' | 'warning' | 'info';
  }): Promise<boolean> {
    this.title = options.title || 'Confirm Action';
    this.message = options.message;
    this.confirmText = options.confirmText || 'Confirm';
    this.cancelText = options.cancelText || 'Cancel';
    this.variant = options.variant || 'danger';
    this.open = true;

    return new Promise<boolean>((resolve) => {
      this.resolvePromise = resolve;
    });
  }

  private handleConfirm() {
    this.open = false;
    if (this.resolvePromise) {
      this.resolvePromise(true);
    }
    this.resolvePromise = undefined;
  }

  private handleCancel() {
    this.open = false;
    if (this.resolvePromise) {
      this.resolvePromise(false);
    }
    this.resolvePromise = undefined;
  }

  override render() {
    if (!this.open) return html``;

    const confirmClass = `btn btn-${this.variant === 'info' ? 'info' : this.variant === 'warning' ? 'warning' : 'danger'}`;

    return html`
      <div class="modal" @click=${(e: Event) => e.target === e.currentTarget && this.handleCancel()}>
        <div class="modal-content">
          <div class="modal-header">
            <h2>${this.title}</h2>
            <button class="close-btn" @click=${this.handleCancel} type="button">&times;</button>
          </div>
          <div class="modal-body">
            <p class="message">${this.message}</p>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click=${this.handleCancel}>
              ${this.cancelText}
            </button>
            <button type="button" class="${confirmClass}" @click=${this.handleConfirm}>
              ${this.confirmText}
            </button>
          </div>
        </div>
      </div>
    `;
  }
}

// Singleton instance for easy access
let confirmationInstance: ConfirmationModal | null = null;

export async function confirmAction(options: {
  title?: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  variant?: 'danger' | 'warning' | 'info';
}): Promise<boolean> {
  if (!confirmationInstance) {
    confirmationInstance = document.createElement('confirmation-modal') as ConfirmationModal;
    document.body.appendChild(confirmationInstance);
  }
  return confirmationInstance.confirm(options);
}

declare global {
  interface HTMLElementTagNameMap {
    'confirmation-modal': ConfirmationModal;
  }
}
