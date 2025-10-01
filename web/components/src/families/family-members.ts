/**
 * FamilyMembers Lit Component
 *
 * Displays grid of family members with actions
 */

import { LitElement, html, css } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FamilyMember, FamilyService } from './family-service.js';
import { buttonStyles } from '../common/shared-styles.js';
import { errorHandler } from '../common/error-handler.js';
import { familyContext } from './family-context.js';
import { confirmAction } from '../common/confirmation-modal.js';
import { showToast } from '../common/toast-notification.js';
import './family-member-modal.js';

@customElement('family-members-grid')
export class FamilyMembersGrid extends LitElement {
  @state()
  private members: FamilyMember[] = [];

  @state()
  private isLoading = true;

  @state()
  private errorMessage = '';

  private familyService = new FamilyService({ apiBaseUrl: '/api/v1', csrfToken: '' });

  static override styles = [
    buttonStyles,
    css`
      :host {
        display: block;
      }

      .loading-container, .error-container, .empty-members {
        text-align: center;
        padding: 40px 20px;
        color: #6c757d;
      }

      .error-message {
        color: #dc3545;
        margin-bottom: 16px;
      }

      .members-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
        gap: 20px;
        padding: 20px 0;
      }

      .member-card {
        background: white;
        border: 1px solid #e5e7eb;
        border-radius: 8px;
        padding: 20px;
        box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
        transition: box-shadow 0.2s;
      }

      .member-card:hover {
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
      }

      .member-header {
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
        margin-bottom: 16px;
      }

      .member-avatar {
        width: 60px;
        height: 60px;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        background: #f3f4f6;
        font-size: 32px;
        overflow: hidden;
      }

      .member-avatar-img {
        width: 100%;
        height: 100%;
        object-fit: cover;
      }

      .member-actions {
        display: flex;
        gap: 8px;
      }

      .member-action-btn {
        background: none;
        border: 1px solid #d1d5db;
        border-radius: 4px;
        padding: 4px 8px;
        cursor: pointer;
        font-size: 16px;
        color: #6b7280;
        transition: all 0.2s;
      }

      .member-action-btn:hover {
        background: #f3f4f6;
        border-color: #9ca3af;
        color: #374151;
      }

      .member-info {
        text-align: center;
      }

      .member-name {
        margin: 0 0 4px 0;
        font-size: 18px;
        font-weight: 600;
        color: #374151;
      }

      .member-nickname {
        margin: 0 0 8px 0;
        font-size: 14px;
        color: #6b7280;
        font-style: italic;
      }

      .member-age, .member-email {
        margin: 4px 0;
        font-size: 14px;
        color: #6b7280;
      }

      .member-type {
        display: inline-block;
        margin-top: 8px;
        padding: 4px 12px;
        border-radius: 12px;
        font-size: 12px;
        font-weight: 500;
        text-transform: capitalize;
      }

      .member-type.member-type-adult {
        background: #dbeafe;
        color: #1e40af;
      }

      .member-type.member-type-child {
        background: #fef3c7;
        color: #92400e;
      }

      .member-type.member-type-pet {
        background: #d1fae5;
        color: #065f46;
      }

      .empty-subtitle {
        margin-top: 8px;
        font-size: 14px;
      }
    `
  ];

  override async connectedCallback() {
    super.connectedCallback();
    await this.loadMembers();
  }

  private async loadMembers() {
    this.isLoading = true;
    this.errorMessage = '';

    const result = await errorHandler.handleAsync(
      async () => {
        const familyId = await familyContext.getFamilyId();
        if (!familyId) {
          throw new Error('No family ID found. Please log in again.');
        }
        return await this.familyService.listFamilyMembers(familyId);
      },
      { component: 'FamilyMembers', operation: 'loadMembers' },
      []
    );

    this.members = result || [];
    this.isLoading = false;
  }

  private getMemberIcon(memberType: string): string {
    switch (memberType) {
      case 'adult':
        return '👨‍👩‍👧‍👦';
      case 'child':
        return '👶';
      case 'pet':
        return '🐕';
      default:
        return '👤';
    }
  }

  private handleEdit(member: FamilyMember) {
    const modal = this.shadowRoot?.querySelector('family-member-modal') as any;
    if (modal) {
      modal.show(member);
    }
  }

  private async handleDelete(member: FamilyMember) {
    const confirmed = await confirmAction({
      title: 'Delete Family Member',
      message: `Are you sure you want to delete ${member.name}? This action cannot be undone.`,
      confirmText: 'Delete',
      cancelText: 'Cancel',
      variant: 'danger'
    });

    if (!confirmed) return;

    const success = await errorHandler.handleAsync(
      async () => {
        const familyId = await familyContext.getFamilyId();
        if (!familyId) {
          throw new Error('No family ID found');
        }
        await this.familyService.deleteFamilyMember(familyId, member.id);
        await this.loadMembers();
        return true;
      },
      { component: 'FamilyMembers', operation: 'deleteMember' }
    );

    if (success) {
      showToast(`${member.name} has been deleted`, 'success');
      this.dispatchEvent(new CustomEvent('member-deleted', {
        detail: { id: member.id },
        bubbles: true
      }));
    }
  }

  private async handleModalSave(e: CustomEvent) {
    const { data, memberId } = e.detail;

    const success = await errorHandler.handleAsync(
      async () => {
        const familyId = await familyContext.getFamilyId();
        if (!familyId) {
          throw new Error('No family ID found');
        }

        if (memberId) {
          await this.familyService.updateFamilyMember(familyId, memberId, data);
        } else {
          await this.familyService.createFamilyMember(familyId, data);
        }
        await this.loadMembers();
        return true;
      },
      { component: 'FamilyMembers', operation: memberId ? 'updateMember' : 'createMember' }
    );

    if (success) {
      const message = memberId ? `${data.name} has been updated` : `${data.name} has been added to your family`;
      showToast(message, 'success');
      this.dispatchEvent(new CustomEvent(memberId ? 'member-updated' : 'member-added', {
        detail: { memberId },
        bubbles: true
      }));
    }
  }

  showAddModal() {
    const modal = this.shadowRoot?.querySelector('family-member-modal') as any;
    if (modal) {
      modal.show();
    }
  }

  async refresh() {
    await this.loadMembers();
  }

  private renderLoading() {
    return html`
      <div class="loading-container">
        <p>Loading family members...</p>
      </div>
    `;
  }

  private renderError() {
    return html`
      <div class="error-container">
        <p class="error-message">${this.errorMessage}</p>
        <button class="btn btn-secondary" @click=${this.loadMembers}>Retry</button>
      </div>
    `;
  }

  private renderEmpty() {
    return html`
      <div class="empty-members">
        <p>No family members yet</p>
        <p class="empty-subtitle">Add your first family member to get started</p>
      </div>
    `;
  }

  private renderMemberCard(member: FamilyMember) {
    const memberTypeClass = `member-type-${member.member_type}`;
    const memberIcon = this.getMemberIcon(member.member_type);

    return html`
      <div class="member-card" data-member-id="${member.id}">
        <div class="member-header">
          <div class="member-avatar">
            ${member.avatar_url
              ? html`<img src="${member.avatar_url}" alt="${member.name}" class="member-avatar-img" />`
              : html`<span class="member-icon">${memberIcon}</span>`}
          </div>
          <div class="member-actions">
            <button
              class="member-action-btn"
              @click=${() => this.handleEdit(member)}
              title="Edit"
            >
              ✎
            </button>
            <button
              class="member-action-btn"
              @click=${() => this.handleDelete(member)}
              title="Delete"
            >
              ×
            </button>
          </div>
        </div>
        <div class="member-info">
          <h4 class="member-name">${member.name}</h4>
          ${member.nickname ? html`<p class="member-nickname">"${member.nickname}"</p>` : ''}
          ${member.age ? html`<p class="member-age">Age: ${member.age}</p>` : ''}
          ${member.user?.email ? html`<p class="member-email">${member.user.email}</p>` : ''}
          <span class="member-type ${memberTypeClass}">${member.member_type}</span>
        </div>
      </div>
    `;
  }

  override render() {
    if (this.isLoading) return this.renderLoading();
    if (this.errorMessage) return this.renderError();
    if (this.members.length === 0) return this.renderEmpty();

    return html`
      <div class="members-grid">
        ${this.members.map(member => this.renderMemberCard(member))}
      </div>
      <family-member-modal @save=${this.handleModalSave}></family-member-modal>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'family-members-grid': FamilyMembersGrid;
  }
}
