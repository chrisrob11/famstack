/**
 * IntegrationCard Component
 *
 * Role: Renders a single integration as a card with status and actions
 * Responsibilities:
 * - Display integration information (name, provider, status, description)
 * - Show visual status indicators and category icons
 * - Render action buttons through IntegrationActions component
 * - Forward action events to parent components
 * - Provide consistent card styling and layout
 */

import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import type { Integration } from './integration-types.js';
import {
  getCategoryIcon,
  getProviderLabel,
  STATUS_LABELS,
} from './integration-types.js';
import { buttonStyles } from '../common/shared-styles.js';

@customElement('integration-card')
export class IntegrationCard extends LitElement {
  @property({ type: Object })
  integration!: Integration;

  static override styles = [
    buttonStyles,
    css`
      :host {
        display: block;
      }

      .integration-card {
        background: #fff;
        border: 1px solid #e1e5e9;
        border-radius: 8px;
        padding: 20px;
        box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
        transition: box-shadow 0.2s ease;
      }

      .integration-card:hover {
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
      }

      .integration-header {
        display: flex;
        align-items: center;
        gap: 12px;
        margin-bottom: 16px;
      }

      .integration-icon {
        width: 40px;
        height: 40px;
        display: flex;
        align-items: center;
        justify-content: center;
        border-radius: 8px;
        font-size: 20px;
        background: #f8f9fa;
      }

      .integration-icon.calendar {
        background: #e3f2fd;
        color: #1976d2;
      }

      .integration-icon.communication {
        background: #f3e5f5;
        color: #7b1fa2;
      }

      .integration-icon.productivity {
        background: #e8f5e8;
        color: #388e3c;
      }

      .integration-info h3 {
        margin: 0 0 4px 0;
        font-size: 16px;
        font-weight: 600;
        color: #333;
      }

      .integration-info .provider {
        margin: 0;
        font-size: 12px;
        color: #6c757d;
        text-transform: uppercase;
        font-weight: 500;
      }

      .integration-status {
        display: flex;
        align-items: center;
        gap: 8px;
        margin-bottom: 12px;
        font-size: 12px;
        font-weight: 500;
      }

      .status-indicator {
        width: 8px;
        height: 8px;
        border-radius: 50%;
      }

      .status-indicator.connected {
        background: #28a745;
      }

      .status-indicator.disconnected {
        background: #dc3545;
      }

      .status-indicator.pending {
        background: #ffc107;
      }

      .status-indicator.error {
        background: #fd7e14;
      }

      .integration-description {
        margin: 0 0 16px 0;
        font-size: 14px;
        color: #6c757d;
        line-height: 1.4;
      }
    `
  ];

  private handleActionClick(action: string) {
    this.dispatchEvent(
      new CustomEvent('integration-action', {
        detail: { action, integrationId: this.integration.id },
        bubbles: true,
      })
    );
  }

  override render() {
    return html`
      <div class="integration-card">
        <div class="integration-header">
          <div class="integration-icon ${this.integration.integration_type}">
            ${getCategoryIcon(this.integration.integration_type)}
          </div>
          <div class="integration-info">
            <h3>${this.integration.display_name}</h3>
            <p class="provider">${getProviderLabel(this.integration.provider)}</p>
          </div>
        </div>
        <div class="integration-status">
          <div class="status-indicator ${this.integration.status}"></div>
          <span>${STATUS_LABELS[this.integration.status] || this.integration.status}</span>
        </div>
        ${this.integration.description
          ? html`<p class="integration-description">${this.integration.description}</p>`
          : ''}
        <integration-actions
          .integration=${this.integration}
          @action-click=${(e: CustomEvent) => this.handleActionClick(e.detail.action)}>
        </integration-actions>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'integration-card': IntegrationCard;
  }
}