import { ComponentConfig } from '../common/types.js';
import { FamilyMember, CreateFamilyMemberRequest } from './family-service.js';

export type FamilyMemberFormData = CreateFamilyMemberRequest;

export interface FamilyMemberModalOptions {
  onSave: (data: FamilyMemberFormData, memberId?: string) => Promise<void>;
  onCancel?: () => void;
}

export class FamilyMemberModal {
  private element!: HTMLElement;
  private backdrop!: HTMLElement;
  private currentMember: FamilyMember | undefined;
  private isSubmitting: boolean = false;
  private onSave: (data: FamilyMemberFormData, memberId?: string) => Promise<void>;
  private onCancel: (() => void) | undefined;
  private boundHandleClick?: (e: Event) => void;
  private boundHandleSubmit?: (e: Event) => void;

  constructor(container: HTMLElement, config: ComponentConfig, options: FamilyMemberModalOptions) {
    this.onSave = options.onSave;
    this.onCancel = options.onCancel;

    this.createElement(container);
    this.attachEventListeners();
  }

  private createElement(container: HTMLElement): void {
    // Create modal
    this.element = document.createElement('div');
    this.element.className = 'modal';
    this.element.id = 'family-member-modal';
    this.element.style.display = 'none';
    this.element.innerHTML = this.getModalHTML();

    // Create backdrop
    this.backdrop = document.createElement('div');
    this.backdrop.className = 'modal-backdrop';
    this.backdrop.style.display = 'none';

    container.appendChild(this.element);
    container.appendChild(this.backdrop);
  }

  private getModalHTML(): string {
    return `
      <div class="modal-content">
        <div class="modal-header">
          <h2 id="modal-title">Add Family Member</h2>
          <button class="modal-close" data-action="close-modal" type="button">×</button>
        </div>
        <form id="family-member-form">
          <input type="hidden" id="member-id" name="id">

          <div class="form-group">
            <label for="member-name">Name *</label>
            <input type="text" id="member-name" name="name" required
                   placeholder="Enter member name">
          </div>

          <div class="form-group">
            <label for="member-nickname">Nickname</label>
            <input type="text" id="member-nickname" name="nickname"
                   placeholder="Enter nickname (optional)">
          </div>

          <div class="form-group">
            <label for="member-type">Member Type *</label>
            <select id="member-type" name="member_type" required>
              <option value="">Select member type</option>
              <option value="adult">Adult</option>
              <option value="child">Child</option>
              <option value="pet">Pet</option>
            </select>
          </div>

          <div class="form-group">
            <label for="member-age">Age</label>
            <input type="number" id="member-age" name="age" min="0" max="150"
                   placeholder="Enter age (optional)">
          </div>

          <div class="form-group">
            <label for="member-avatar-url">Avatar URL</label>
            <input type="url" id="member-avatar-url" name="avatar_url"
                   placeholder="Enter avatar URL (optional)">
          </div>

          <div class="modal-actions">
            <button type="button" class="btn btn-secondary" data-action="close-modal">Cancel</button>
            <button type="submit" class="btn btn-primary" id="submit-btn">Add Member</button>
          </div>
        </form>
      </div>
    `;
  }

  private attachEventListeners(): void {
    this.boundHandleClick = this.handleClick.bind(this);
    this.boundHandleSubmit = this.handleSubmit.bind(this);

    this.element.addEventListener('click', this.boundHandleClick);
    this.element.addEventListener('submit', this.boundHandleSubmit);
  }

  public showCreate(): void {
    this.currentMember = undefined;

    const title = this.element.querySelector('#modal-title') as HTMLElement;
    const submitBtn = this.element.querySelector('#submit-btn') as HTMLButtonElement;

    title.textContent = 'Add Family Member';
    submitBtn.textContent = 'Add Member';

    this.resetForm();
    this.show();
  }

  public showEdit(member: FamilyMember): void {
    this.currentMember = member;

    const title = this.element.querySelector('#modal-title') as HTMLElement;
    const submitBtn = this.element.querySelector('#submit-btn') as HTMLButtonElement;

    title.textContent = 'Edit Family Member';
    submitBtn.textContent = 'Update Member';

    this.populateForm(member);
    this.show();
  }

  private show(): void {
    this.element.style.display = 'flex';
    this.backdrop.style.display = 'block';

    const nameInput = this.element.querySelector('#member-name') as HTMLInputElement;
    nameInput?.focus();
  }

  public hide(): void {
    this.element.style.display = 'none';
    this.backdrop.style.display = 'none';
    this.resetForm();
  }

  private resetForm(): void {
    const form = this.element.querySelector('#family-member-form') as HTMLFormElement;
    form.reset();

    // Clear any error messages
    const errors = form.querySelectorAll('.field-error');
    errors.forEach(error => error.remove());
  }

  private populateForm(member: FamilyMember): void {
    const form = this.element.querySelector('#family-member-form') as HTMLFormElement;

    (form.querySelector('#member-id') as HTMLInputElement).value = member.id;
    (form.querySelector('#member-name') as HTMLInputElement).value = member.name;
    (form.querySelector('#member-nickname') as HTMLInputElement).value = member.nickname || '';
    (form.querySelector('#member-type') as HTMLSelectElement).value = member.member_type;
    (form.querySelector('#member-age') as HTMLInputElement).value = member.age?.toString() || '';
    (form.querySelector('#member-avatar-url') as HTMLInputElement).value = member.avatar_url || '';
  }

  private handleClick(e: Event): void {
    const target = e.target as HTMLElement;
    const action = target.getAttribute('data-action');

    if (action === 'close-modal') {
      e.preventDefault();
      this.hide();
      this.onCancel?.();
    }
  }

  private handleSubmit(e: Event): void {
    e.preventDefault();
    if (this.isSubmitting) return;

    const form = e.target as HTMLFormElement;
    this.handleFormSubmit(form);
  }

  private async handleFormSubmit(form: HTMLFormElement): Promise<void> {
    if (this.isSubmitting) return;

    this.isSubmitting = true;

    try {
      const formData = this.getFormData(form);
      const memberId = this.currentMember?.id;

      await this.onSave(formData, memberId);
      this.hide();
    } catch (error) {
      this.showFormError(
        form,
        error instanceof Error ? error.message : 'Failed to save family member'
      );
    } finally {
      this.isSubmitting = false;
    }
  }

  private getFormData(form: HTMLFormElement): FamilyMemberFormData {
    const formData = new FormData(form);

    const name = formData.get('name') as string;
    const nickname = formData.get('nickname') as string;
    const memberType = formData.get('member_type') as string;
    const age = formData.get('age') as string;
    const avatarUrl = formData.get('avatar_url') as string;

    if (!name.trim()) {
      throw new Error('Name is required');
    }

    if (!memberType) {
      throw new Error('Member type is required');
    }

    const result: FamilyMemberFormData = {
      name: name.trim(),
      member_type: memberType as 'adult' | 'child' | 'pet',
    };

    if (nickname.trim()) {
      result.nickname = nickname.trim();
    }

    if (age.trim()) {
      const ageNum = parseInt(age.trim(), 10);
      if (!isNaN(ageNum) && ageNum >= 0) {
        result.age = ageNum;
      }
    }

    if (avatarUrl.trim()) {
      result.avatar_url = avatarUrl.trim();
    }

    return result;
  }

  private showFormError(form: HTMLFormElement, message: string): void {
    // Clear existing errors
    const existingErrors = form.querySelectorAll('.field-error');
    existingErrors.forEach(error => error.remove());

    // Add new error
    const errorDiv = document.createElement('div');
    errorDiv.className = 'field-error form-error-general';
    errorDiv.textContent = message;
    errorDiv.style.cssText = `
      color: #e53e3e;
      font-size: 0.875rem;
      margin-top: 0.5rem;
      padding: 0.5rem;
      background: #fed7d7;
      border-radius: 4px;
    `;

    form.appendChild(errorDiv);
  }

  public destroy(): void {
    if (this.boundHandleClick) {
      this.element.removeEventListener('click', this.boundHandleClick);
    }
    if (this.boundHandleSubmit) {
      this.element.removeEventListener('submit', this.boundHandleSubmit);
    }

    this.element.remove();
    this.backdrop.remove();
  }
}
