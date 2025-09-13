import { ComponentConfig } from '../common/types.js';
import { FamilyMember, CreateFamilyMemberRequest } from './family-service.js';

export type FamilyMemberFormData = Omit<CreateFamilyMemberRequest, 'family_id'>;

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
            <label for="member-email">Email</label>
            <input type="email" id="member-email" name="email"
                   placeholder="Enter email (optional)">
          </div>

          <div class="form-group">
            <label for="member-role">Role *</label>
            <select id="member-role" name="role" required>
              <option value="">Select role</option>
              <option value="parent">Parent</option>
              <option value="child">Child</option>
            </select>
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
    (form.querySelector('#member-email') as HTMLInputElement).value = member.email || '';
    (form.querySelector('#member-role') as HTMLSelectElement).value = member.role;
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
    const email = formData.get('email') as string;
    const role = formData.get('role') as string;

    if (!name.trim()) {
      throw new Error('Name is required');
    }

    if (!role) {
      throw new Error('Role is required');
    }

    return {
      name: name.trim(),
      email: email.trim() || '',
      role: role,
    };
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
