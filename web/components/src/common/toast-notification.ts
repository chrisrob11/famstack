/**
 * Toast Notification Component
 *
 * Displays temporary success/error/info messages
 */

import { LitElement, html, css } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';

export type ToastType = 'success' | 'error' | 'info' | 'warning';

@customElement('toast-notification')
export class ToastNotification extends LitElement {
  @property({ type: String })
  message = '';

  @property({ type: String })
  type: ToastType = 'info';

  @property({ type: Number })
  duration = 3000;

  @state()
  private visible = false;

  static override styles = css`
    :host {
      position: fixed;
      top: 20px;
      right: 20px;
      z-index: 10000;
      pointer-events: none;
    }

    .toast {
      padding: 16px 24px;
      border-radius: 8px;
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
      display: flex;
      align-items: center;
      gap: 12px;
      min-width: 300px;
      max-width: 500px;
      pointer-events: auto;
      transform: translateX(400px);
      opacity: 0;
      transition: all 0.3s ease-in-out;
    }

    .toast.visible {
      transform: translateX(0);
      opacity: 1;
    }

    .toast.success {
      background: #d1fae5;
      color: #065f46;
      border: 1px solid #6ee7b7;
    }

    .toast.error {
      background: #fecaca;
      color: #991b1b;
      border: 1px solid #f87171;
    }

    .toast.warning {
      background: #fef3c7;
      color: #92400e;
      border: 1px solid #fbbf24;
    }

    .toast.info {
      background: #dbeafe;
      color: #1e40af;
      border: 1px solid #60a5fa;
    }

    .icon {
      font-size: 20px;
      flex-shrink: 0;
    }

    .message {
      flex: 1;
      font-size: 14px;
      font-weight: 500;
    }

    .close-btn {
      background: none;
      border: none;
      font-size: 20px;
      cursor: pointer;
      padding: 0;
      opacity: 0.6;
      transition: opacity 0.2s;
      color: inherit;
    }

    .close-btn:hover {
      opacity: 1;
    }

    @media (max-width: 640px) {
      :host {
        left: 20px;
        right: 20px;
      }

      .toast {
        min-width: unset;
      }
    }
  `;

  private timer: number | undefined = undefined;

  show(message: string, type: ToastType = 'info', duration: number = 3000) {
    this.message = message;
    this.type = type;
    this.duration = duration;
    this.visible = true;

    if (this.timer) {
      window.clearTimeout(this.timer);
    }

    if (duration > 0) {
      this.timer = window.setTimeout(() => {
        this.hide();
      }, duration);
    }
  }

  hide() {
    this.visible = false;
    if (this.timer !== undefined) {
      window.clearTimeout(this.timer);
    }
    this.timer = undefined;
  }

  private getIcon(): string {
    switch (this.type) {
      case 'success':
        return '✓';
      case 'error':
        return '✕';
      case 'warning':
        return '⚠';
      case 'info':
        return 'ℹ';
      default:
        return '';
    }
  }

  override render() {
    return html`
      <div class="toast ${this.type} ${this.visible ? 'visible' : ''}">
        <span class="icon">${this.getIcon()}</span>
        <span class="message">${this.message}</span>
        <button class="close-btn" @click=${this.hide} aria-label="Close">×</button>
      </div>
    `;
  }
}

// Singleton instance for easy access
let toastInstance: ToastNotification | null = null;

export function showToast(message: string, type: ToastType = 'info', duration: number = 3000) {
  if (!toastInstance) {
    toastInstance = document.createElement('toast-notification') as ToastNotification;
    document.body.appendChild(toastInstance);
  }
  toastInstance.show(message, type, duration);
}

declare global {
  interface HTMLElementTagNameMap {
    'toast-notification': ToastNotification;
  }
}
