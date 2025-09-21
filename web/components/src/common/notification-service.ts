/**
 * Notification Service
 * Replaces browser alerts with proper toast notifications
 */

export interface NotificationOptions {
  type?: 'success' | 'error' | 'warning' | 'info';
  duration?: number; // milliseconds, 0 for persistent
  title?: string;
}

export class NotificationService {
  private container: HTMLElement;
  private notifications: Map<string, HTMLElement> = new Map();

  constructor() {
    this.container = this.createContainer();
    this.addStyles();
  }

  private createContainer(): HTMLElement {
    let container = document.getElementById('notification-container');
    if (!container) {
      container = document.createElement('div');
      container.id = 'notification-container';
      container.className = 'notification-container';
      document.body.appendChild(container);
    }
    return container;
  }

  private addStyles(): void {
    if (document.getElementById('notification-styles')) return;

    const styles = document.createElement('style');
    styles.id = 'notification-styles';
    styles.textContent = `
      .notification-container {
        position: fixed;
        top: 1rem;
        right: 1rem;
        z-index: 9999;
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
        max-width: 400px;
        pointer-events: none;
      }

      .notification {
        background: white;
        border-radius: 0.75rem;
        padding: 1rem;
        box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
        border-left: 4px solid;
        display: flex;
        align-items: flex-start;
        gap: 0.75rem;
        pointer-events: auto;
        transform: translateX(100%);
        transition: all 0.3s ease;
        max-width: 100%;
        word-wrap: break-word;
      }

      .notification.show {
        transform: translateX(0);
      }

      .notification.success {
        border-left-color: #10b981;
      }

      .notification.error {
        border-left-color: #ef4444;
      }

      .notification.warning {
        border-left-color: #f59e0b;
      }

      .notification.info {
        border-left-color: #3b82f6;
      }

      .notification-icon {
        font-size: 1.25rem;
        flex-shrink: 0;
        margin-top: 0.125rem;
      }

      .notification-content {
        flex: 1;
        min-width: 0;
      }

      .notification-title {
        font-weight: 600;
        color: #374151;
        margin: 0 0 0.25rem 0;
        font-size: 0.875rem;
      }

      .notification-message {
        color: #6b7280;
        margin: 0;
        font-size: 0.875rem;
        line-height: 1.4;
      }

      .notification-close {
        background: none;
        border: none;
        color: #9ca3af;
        cursor: pointer;
        font-size: 1.25rem;
        padding: 0;
        flex-shrink: 0;
        width: 20px;
        height: 20px;
        display: flex;
        align-items: center;
        justify-content: center;
      }

      .notification-close:hover {
        color: #6b7280;
      }

      @media (max-width: 640px) {
        .notification-container {
          left: 1rem;
          right: 1rem;
          max-width: none;
        }

        .notification {
          margin: 0;
        }
      }
    `;

    document.head.appendChild(styles);
  }

  private getIcon(type: string): string {
    const icons = {
      success: '✅',
      error: '❌',
      warning: '⚠️',
      info: 'ℹ️',
    };
    return icons[type as keyof typeof icons] || 'ℹ️';
  }

  show(message: string, options: NotificationOptions = {}): string {
    const { type = 'info', duration = 5000, title } = options;

    const id = `notification-${Date.now()}-${Math.random().toString(36).substring(2, 11)}`;

    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    notification.innerHTML = `
      <div class="notification-icon">${this.getIcon(type)}</div>
      <div class="notification-content">
        ${title ? `<div class="notification-title">${this.escapeHtml(title)}</div>` : ''}
        <div class="notification-message">${this.escapeHtml(message)}</div>
      </div>
      <button class="notification-close" aria-label="Close notification">&times;</button>
    `;

    // Close button handler
    const closeBtn = notification.querySelector('.notification-close');
    closeBtn?.addEventListener('click', () => this.hide(id));

    this.container.appendChild(notification);
    this.notifications.set(id, notification);

    // Trigger animation
    requestAnimationFrame(() => {
      notification.classList.add('show');
    });

    // Auto-hide if duration is set
    if (duration > 0) {
      setTimeout(() => this.hide(id), duration);
    }

    return id;
  }

  hide(id: string): void {
    const notification = this.notifications.get(id);
    if (!notification) return;

    notification.classList.remove('show');

    setTimeout(() => {
      if (notification.parentNode) {
        notification.parentNode.removeChild(notification);
      }
      this.notifications.delete(id);
    }, 300);
  }

  hideAll(): void {
    for (const id of this.notifications.keys()) {
      this.hide(id);
    }
  }

  success(message: string, title?: string): string {
    return this.show(message, { type: 'success', ...(title && { title }) });
  }

  error(message: string, title?: string): string {
    return this.show(message, { type: 'error', ...(title && { title }) });
  }

  warning(message: string, title?: string): string {
    return this.show(message, { type: 'warning', ...(title && { title }) });
  }

  info(message: string, title?: string): string {
    return this.show(message, { type: 'info', ...(title && { title }) });
  }

  private escapeHtml(text: string): string {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }
}

// Export singleton instance
export const notifications = new NotificationService();
