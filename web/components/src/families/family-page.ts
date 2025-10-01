/**
 * FamilyPage component - handles the family setup page with Lit components
 */

import { LitElement, html, css } from 'lit';
import { customElement } from 'lit/decorators.js';
import { query } from 'lit/decorators.js';
import './family-info.js';
import './family-members.js';

@customElement('family-page')
export class FamilyPage extends LitElement {
  @query('family-members-grid')
  private membersGrid?: any;

  static override styles = css`
    :host {
      display: block;
      padding: 2rem;
      max-width: 1200px;
      margin: 0 auto;
    }

    .setup-header {
      margin-bottom: 2rem;
    }

    .setup-header h2 {
      margin: 0 0 0.5rem 0;
      font-size: 2rem;
      font-weight: 700;
      color: #374151;
    }

    .setup-header p {
      margin: 0;
      color: #6b7280;
      font-size: 1rem;
    }

    .family-members-section {
      background: white;
      border: 1px solid #e5e7eb;
      border-radius: 8px;
      padding: 24px;
    }

    .section-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 20px;
    }

    .section-header h3 {
      margin: 0;
      font-size: 20px;
      font-weight: 600;
      color: #374151;
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

    .btn-secondary {
      background: #f3f4f6;
      color: #374151;
      border: 1px solid #d1d5db;
    }

    .btn-secondary:hover:not(:disabled) {
      background: #e5e7eb;
    }

    @media (max-width: 768px) {
      :host {
        padding: 1rem;
      }

      .section-header {
        flex-direction: column;
        align-items: stretch;
        gap: 12px;
      }
    }
  `;

  private handleAddMember() {
    if (this.membersGrid && this.membersGrid.showAddModal) {
      this.membersGrid.showAddModal();
    }
  }

  private handleFamilyCreated() {
    // Refresh members grid when family is created
    if (this.membersGrid && this.membersGrid.refresh) {
      this.membersGrid.refresh();
    }
  }

  override render() {
    return html`
      <div class="setup-header">
        <h2>Family</h2>
        <p>Manage your family members and settings</p>
      </div>

      <!-- Family Info Section -->
      <family-info @family-created=${this.handleFamilyCreated}></family-info>

      <!-- Family Members Section -->
      <section class="family-members-section">
        <div class="section-header">
          <h3>Family Members</h3>
          <button class="btn btn-secondary" @click=${this.handleAddMember}>
            + Add Member
          </button>
        </div>

        <div class="members-list">
          <family-members-grid></family-members-grid>
        </div>
      </section>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'family-page': FamilyPage;
  }
}

// Export for compatibility with page factory
export default FamilyPage;
