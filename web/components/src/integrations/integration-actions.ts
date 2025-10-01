/**
 * IntegrationActions Component
 *
 * Role: Renders contextual action buttons for an integration
 * Responsibilities:
 * - Display appropriate action buttons based on integration status
 * - Handle button click events and dispatch action events
 * - Show Connect button for disconnected/pending OAuth integrations
 * - Show Sync/Test buttons for connected integrations
 * - Always show Configure and Delete buttons
 * - Provide consistent button styling and behavior
 */

import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import type { Integration } from './integration-types.js';
import { INTEGRATION_STATUS, AUTH_METHODS } from '../common/constants.js';
import { buttonStyles } from '../common/shared-styles.js';

@customElement('integration-actions')
export class IntegrationActions extends LitElement {
  @property({ type: Object })
  integration!: Integration;

  static override styles = [
    buttonStyles,
    css`
      :host {
        display: block;
      }

      .integration-actions {
        display: flex;
        gap: 8px;
        flex-wrap: wrap;
      }

      .btn {
        padding: 6px 12px;
        font-size: 12px;
      }
    `,
  ];

  private handleActionClick(action: string) {
    this.dispatchEvent(
      new CustomEvent('action-click', {
        detail: { action },
        bubbles: true,
      })
    );
  }

  private renderActionButtons() {
    const actions = [];

    // Connect button for disconnected or pending OAuth integrations
    if (
      this.integration.status === INTEGRATION_STATUS.DISCONNECTED ||
      (this.integration.status === INTEGRATION_STATUS.PENDING &&
        this.integration.auth_method === AUTH_METHODS.OAUTH2)
    ) {
      actions.push(
        html`<button class="btn btn-secondary" @click=${() => this.handleActionClick('connect')}>
          Connect
        </button>`
      );
    }

    // Sync and Test buttons for connected integrations
    if (this.integration.status === INTEGRATION_STATUS.CONNECTED) {
      actions.push(
        html`<button class="btn btn-secondary" @click=${() => this.handleActionClick('sync')}>
          Sync
        </button>`
      );
      actions.push(
        html`<button class="btn btn-secondary" @click=${() => this.handleActionClick('test')}>
          Test
        </button>`
      );
    }

    // Configure button (always available)
    actions.push(
      html`<button class="btn btn-secondary" @click=${() => this.handleActionClick('configure')}>
        Configure
      </button>`
    );

    // Delete button (always available)
    actions.push(
      html`<button class="btn btn-danger" @click=${() => this.handleActionClick('delete')}>
        Delete
      </button>`
    );

    return actions;
  }

  override render() {
    return html` <div class="integration-actions">${this.renderActionButtons()}</div> `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'integration-actions': IntegrationActions;
  }
}
