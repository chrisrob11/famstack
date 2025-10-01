/**
 * FamilyMemberModal Lit Component
 *
 * Modal for adding/editing family members
 */

import { LitElement, html, css } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { FamilyMember, CreateFamilyMemberRequest } from './family-service.js';
import { modalStyles, buttonStyles, formStyles } from '../common/shared-styles.js';

export type FamilyMemberFormData = CreateFamilyMemberRequest;

@customElement('family-member-modal')
export class FamilyMemberModal extends LitElement {
  @property({ type: Boolean })
  open = false;

  @property({ type: Object })
  member: FamilyMember | null = null;

  @state()
  private isSubmitting = false;

  @state()
  private errorMessage = '';

  static override styles = [
    buttonStyles,
    modalStyles,
    formStyles,
    css`
      :host {
        display: none;
      }

      :host([open]) {
        display: block;
      }

      .form-error-general {
        color: #e53e3e;
        font-size: 0.875rem;
        margin-top: 0.5rem;
        padding: 0.5rem;
        background: #fed7d7;
        border-radius: 4px;
      }

      .modal-actions {
        display: flex;
        justify-content: flex-end;
        gap: 12px;
        margin-top: 24px;
      }
    `
  ];

  override updated(changedProperties: Map<string, any>) {
    if (changedProperties.has('open') && this.open) {
      // Focus the name input when modal opens
      setTimeout(() => {
        const nameInput = this.shadowRoot?.querySelector('#member-name') as HTMLInputElement;
        nameInput?.focus();
      }, 100);
    }
  }

  private close() {
    this.open = false;
    this.member = null;
    this.errorMessage = '';
    this.dispatchEvent(new CustomEvent('close', { bubbles: true }));
  }

  private async handleSubmit(e: Event) {
    e.preventDefault();
    if (this.isSubmitting) return;

    const form = e.target as HTMLFormElement;
    const formData = new FormData(form);

    const name = (formData.get('name') as string).trim();
    const nickname = (formData.get('nickname') as string).trim();
    const memberType = formData.get('member_type') as string;
    const age = formData.get('age') as string;
    const avatarUrl = (formData.get('avatar_url') as string).trim();

    if (!name) {
      this.errorMessage = 'Name is required';
      return;
    }

    if (!memberType) {
      this.errorMessage = 'Member type is required';
      return;
    }

    const data: FamilyMemberFormData = {
      name,
      member_type: memberType as 'adult' | 'child' | 'pet',
    };

    if (nickname) data.nickname = nickname;
    if (age) {
      const ageNum = parseInt(age, 10);
      if (!isNaN(ageNum) && ageNum >= 0) {
        data.age = ageNum;
      }
    }
    if (avatarUrl) data.avatar_url = avatarUrl;

    this.isSubmitting = true;
    this.errorMessage = '';

    try {
      this.dispatchEvent(new CustomEvent('save', {
        detail: { data, memberId: this.member?.id },
        bubbles: true
      }));
      this.close();
    } catch (error) {
      this.errorMessage = error instanceof Error ? error.message : 'Failed to save family member';
    } finally {
      this.isSubmitting = false;
    }
  }

  show(member?: FamilyMember) {
    this.member = member || null;
    this.open = true;
    this.errorMessage = '';
  }

  hide() {
    this.close();
  }

  override render() {
    if (!this.open) return html``;

    const isEdit = !!this.member;
    const title = isEdit ? 'Edit Family Member' : 'Add Family Member';
    const submitText = isEdit ? 'Update Member' : 'Add Member';

    return html`
      <div class="modal" @click=${(e: Event) => e.target === e.currentTarget && this.close()}>
        <div class="modal-content">
          <div class="modal-header">
            <h2>${title}</h2>
            <button class="close-btn" @click=${this.close} type="button">&times;</button>
          </div>
          <div class="modal-body">
            <form @submit=${this.handleSubmit}>
              <div class="form-group">
                <label for="member-name">Name *</label>
                <input
                  type="text"
                  id="member-name"
                  name="name"
                  .value=${this.member?.name || ''}
                  required
                  placeholder="Enter member name"
                />
              </div>

              <div class="form-group">
                <label for="member-nickname">Nickname</label>
                <input
                  type="text"
                  id="member-nickname"
                  name="nickname"
                  .value=${this.member?.nickname || ''}
                  placeholder="Enter nickname (optional)"
                />
              </div>

              <div class="form-group">
                <label for="member-type">Member Type *</label>
                <select
                  id="member-type"
                  name="member_type"
                  .value=${this.member?.member_type || ''}
                  required
                >
                  <option value="">Select member type</option>
                  <option value="adult">Adult</option>
                  <option value="child">Child</option>
                  <option value="pet">Pet</option>
                </select>
              </div>

              <div class="form-group">
                <label for="member-age">Age</label>
                <input
                  type="number"
                  id="member-age"
                  name="age"
                  .value=${this.member?.age?.toString() || ''}
                  min="0"
                  max="150"
                  placeholder="Enter age (optional)"
                />
              </div>

              <div class="form-group">
                <label for="member-avatar-url">Avatar URL</label>
                <input
                  type="url"
                  id="member-avatar-url"
                  name="avatar_url"
                  .value=${this.member?.avatar_url || ''}
                  placeholder="Enter avatar URL (optional)"
                />
              </div>

              ${this.errorMessage
                ? html`<div class="form-error-general">${this.errorMessage}</div>`
                : ''}

              <div class="modal-actions">
                <button type="button" class="btn btn-secondary" @click=${this.close} ?disabled=${this.isSubmitting}>
                  Cancel
                </button>
                <button type="submit" class="btn btn-primary" ?disabled=${this.isSubmitting}>
                  ${this.isSubmitting ? (isEdit ? 'Updating...' : 'Adding...') : submitText}
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'family-member-modal': FamilyMemberModal;
  }
}
