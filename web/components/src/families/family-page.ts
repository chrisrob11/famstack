import { BasePage } from '../pages/base-page.js';
import { FamilyMembers } from './family-members.js';
import { ComponentConfig } from '../common/types.js';

/**
 * FamilyPage component - handles the family setup page
 */
export class FamilyPage extends BasePage {
  private familyMembers?: FamilyMembers;
  private boundHandleSuccess?: (e: Event) => void;

  constructor(container: HTMLElement, config: ComponentConfig) {
    super(container, config, 'family');
  }

  async init(): Promise<void> {
    try {
      this.showLoading('Loading family setup...');

      // Create the page structure
      this.container.innerHTML = this.renderPageContent();

      // Initialize the family members component
      const membersContainer = this.container.querySelector(
        '#family-members-container'
      ) as HTMLElement;
      if (membersContainer) {
        this.familyMembers = new FamilyMembers(membersContainer, this.config);
      }

      // Set up event handlers
      this.setupEventHandlers();
    } catch (error) {
      this.showError('Failed to load family setup page');
    }
  }

  private renderPageContent(): string {
    return `
      <div class="family-page">
        <div class="setup-header">
          <h2>Family</h2>
          <p>Manage your family members and settings</p>
        </div>

        <!-- Family Selection Section -->
        <section class="family-info-section">
          <h3>Current Family</h3>
          <form class="family-form" data-form="create-family">
            <div class="form-group">
              <label for="family-name">Family Name</label>
              <input type="text" id="family-name" name="name" placeholder="Enter your family name" required>
            </div>
            <button type="submit" class="btn btn-primary">Create New Family</button>
          </form>
          <div id="family-info-result"></div>
        </section>

        <!-- Family Members Section -->
        <section class="family-members-section">
          <div class="section-header">
            <h3>Family Members</h3>
            <button class="btn btn-secondary" data-action="show-add-member">+ Add Member</button>
          </div>
          
          <div class="members-list">
            <div id="family-members-container"></div>
          </div>
        </section>

        <!-- Add Member Modal -->
        <div class="modal" id="add-member-modal" style="display: none;">
          <div class="modal-content">
            <div class="modal-header">
              <h3>Add Family Member</h3>
              <button class="modal-close" data-action="hide-add-member">&times;</button>
            </div>
            <form class="member-form" data-form="add-member">
              <div class="form-group">
                <label for="member-name">Name</label>
                <input type="text" id="member-name" name="name" placeholder="Enter member name" required>
              </div>
              <div class="form-group">
                <label for="member-nickname">Nickname</label>
                <input type="text" id="member-nickname" name="nickname" placeholder="Enter nickname (optional)">
              </div>
              <div class="form-group">
                <label for="member-type">Member Type</label>
                <select id="member-type" name="member_type" required>
                  <option value="adult">Adult</option>
                  <option value="child">Child</option>
                  <option value="pet">Pet</option>
                </select>
              </div>
              <div class="form-group">
                <label for="member-age">Age</label>
                <input type="number" id="member-age" name="age" placeholder="Enter age (optional)" min="0" max="150">
              </div>
              <div class="form-group">
                <label for="member-avatar-url">Avatar URL</label>
                <input type="url" id="member-avatar-url" name="avatar_url" placeholder="Enter avatar URL (optional)">
              </div>
              <div class="form-actions">
                <button type="button" class="btn btn-secondary" data-action="hide-add-member">Cancel</button>
                <button type="submit" class="btn btn-primary">Add Member</button>
              </div>
            </form>
            <div id="add-member-result"></div>
          </div>
        </div>
      </div>
    `;
  }

  private setupEventHandlers(): void {
    // Bind methods to preserve `this` context
    this.boundHandleSuccess = this.handleFormSuccess.bind(this);

    // Add event listeners
    this.container.addEventListener('click', this.handleClick.bind(this));
    this.container.addEventListener('submit', this.handleSubmit.bind(this));

    // Listen for successful form submissions
    this.container.addEventListener('form-success', this.boundHandleSuccess);
  }

  private handleClick(e: Event): void {
    const target = e.target as HTMLElement;
    const action = target.getAttribute('data-action');

    switch (action) {
      case 'show-add-member':
        this.showAddMemberModal();
        break;
      case 'hide-add-member':
        this.hideAddMemberModal();
        break;
    }
  }

  private async handleSubmit(e: Event): Promise<void> {
    e.preventDefault();
    const form = e.target as HTMLFormElement;
    const formType = form.getAttribute('data-form');

    switch (formType) {
      case 'create-family':
        await this.handleCreateFamily(form);
        break;
      case 'add-member':
        await this.handleAddMember(form);
        break;
    }
  }

  private async handleCreateFamily(form: HTMLFormElement): Promise<void> {
    const formData = new FormData(form);
    const familyData = {
      name: formData.get('name') as string,
    };

    try {
      const response = await fetch(`${this.config.apiBaseUrl}/families`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': this.config.csrfToken,
        },
        body: JSON.stringify(familyData),
      });

      if (response.ok) {
        const family = await response.json();
        this.showSuccessMessage(`Family "${family.name}" created successfully!`);
        form.reset();
      } else {
        throw new Error('Failed to create family');
      }
    } catch (error) {
      this.showErrorMessage('Failed to create family. Please try again.');
    }
  }

  private async handleAddMember(form: HTMLFormElement): Promise<void> {
    const formData = new FormData(form);
    const memberData: any = {
      name: formData.get('name') as string,
      member_type: formData.get('member_type') as string,
    };

    // Add optional fields if provided
    const nickname = formData.get('nickname') as string;
    const age = formData.get('age') as string;
    const avatarUrl = formData.get('avatar_url') as string;

    if (nickname.trim()) {
      memberData.nickname = nickname.trim();
    }
    if (age.trim()) {
      const ageNum = parseInt(age.trim(), 10);
      if (!isNaN(ageNum) && ageNum >= 0) {
        memberData.age = ageNum;
      }
    }
    if (avatarUrl.trim()) {
      memberData.avatar_url = avatarUrl.trim();
    }

    try {
      // Get family ID from session
      const familyId = await this.getCurrentFamilyId();
      if (!familyId) {
        this.showErrorMessage('No family ID found. Please log in again.');
        return;
      }

      const response = await fetch(`/api/families/${familyId}/members`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(memberData),
      });

      if (response.ok) {
        const member = await response.json();
        this.showSuccessMessage(`Member "${member.name}" added successfully!`);
        this.hideAddMemberModal();
        form.reset();

        // Refresh the family members component
        if (this.familyMembers) {
          await this.familyMembers.refresh();
        }

        // Dispatch success event
        this.container.dispatchEvent(
          new CustomEvent('form-success', {
            detail: { type: 'add-member', data: member },
          })
        );
      } else {
        throw new Error('Failed to add member');
      }
    } catch (error) {
      this.showErrorMessage('Failed to add member. Please try again.');
    }
  }

  private showAddMemberModal(): void {
    const modal = this.container.querySelector('#add-member-modal') as HTMLElement;
    if (modal) {
      modal.style.display = 'flex';
      const nameInput = modal.querySelector('#member-name') as HTMLInputElement;
      nameInput?.focus();
    }
  }

  private hideAddMemberModal(): void {
    const modal = this.container.querySelector('#add-member-modal') as HTMLElement;
    if (modal) {
      modal.style.display = 'none';
      const form = modal.querySelector('.member-form') as HTMLFormElement;
      form?.reset();
      this.clearMessages();
    }
  }

  private handleFormSuccess(e: Event): void {
    // Handle successful form submissions if needed
    const customEvent = e as CustomEvent;
    void customEvent.detail;
  }

  private showSuccessMessage(message: string): void {
    const resultDiv = this.container.querySelector('#family-info-result, #add-member-result');
    if (resultDiv) {
      resultDiv.innerHTML = `<div class="success-message">${message}</div>`;
      setTimeout(() => this.clearMessages(), 3000);
    }
  }

  private showErrorMessage(message: string): void {
    const resultDiv = this.container.querySelector('#family-info-result, #add-member-result');
    if (resultDiv) {
      resultDiv.innerHTML = `<div class="error-message">${message}</div>`;
    }
  }

  private clearMessages(): void {
    const resultDivs = this.container.querySelectorAll('#family-info-result, #add-member-result');
    resultDivs.forEach(div => (div.innerHTML = ''));
  }

  private async getCurrentFamilyId(): Promise<string | null> {
    try {
      const authResponse = await fetch('/auth/me');
      if (!authResponse.ok) {
        return null;
      }
      const sessionData = await authResponse.json();
      return sessionData.session?.family_id || sessionData.user?.family_id || null;
    } catch (error) {
      return null;
    }
  }

  async refresh(): Promise<void> {
    if (this.familyMembers) {
      await this.familyMembers.refresh();
    }
  }

  override destroy(): void {
    if (this.familyMembers) {
      this.familyMembers.destroy();
    }

    // Remove event listeners
    if (this.boundHandleSuccess) {
      this.container.removeEventListener('form-success', this.boundHandleSuccess);
    }

    super.destroy();
  }
}
