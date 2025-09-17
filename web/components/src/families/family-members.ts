import { ComponentConfig } from '../common/types.js';
import { FamilyService, FamilyMember } from './family-service.js';
import { FamilyMemberModal, FamilyMemberFormData } from './family-member-modal.js';
import { ComponentUtils } from '../common/component-utils.js';

export class FamilyMembers {
  private config: ComponentConfig;
  private container: HTMLElement;
  private familyService: FamilyService;
  private familyMemberModal!: FamilyMemberModal;
  private members: FamilyMember[] = [];
  private boundHandleClick?: (e: Event) => void;

  constructor(container: HTMLElement, config: ComponentConfig) {
    this.container = container;
    this.config = config;
    this.familyService = new FamilyService(config);
    this.init();
  }

  private async init(): Promise<void> {
    this.container.className = 'family-members-container';

    // Setup event handling
    this.boundHandleClick = this.handleClick.bind(this);
    this.container.addEventListener('click', this.boundHandleClick);

    // Initialize modal
    this.initializeModal();

    // Listen for refresh events
    this.container.addEventListener('refresh', () => {
      this.loadMembers();
    });

    await this.loadMembers();
  }

  private initializeModal(): void {
    this.familyMemberModal = new FamilyMemberModal(
      this.container.parentElement || document.body,
      this.config,
      {
        onSave: this.handleModalSave.bind(this),
        onCancel: () => {},
      }
    );
  }

  private async loadMembers(): Promise<void> {
    try {
      this.showLoading();
      this.members = await this.familyService.listFamilyMembers('fam1'); // TODO: Get actual family ID
      this.renderMembers();
    } catch (error) {
      this.showErrorInContainer('Failed to load family members');
    }
  }

  private showLoading(): void {
    this.container.innerHTML = `
      <div class="loading-container">
        <div class="loading-spinner"></div>
        <p>Loading family members...</p>
      </div>
    `;
  }

  private showErrorInContainer(message: string): void {
    this.container.innerHTML = `
      <div class="error-container">
        <p class="error-message">${message}</p>
        <button class="btn btn-secondary" onclick="this.parentElement.parentElement.dispatchEvent(new CustomEvent('retry'))">
          Retry
        </button>
      </div>
    `;

    this.container.addEventListener('retry', () => this.loadMembers());
  }

  private renderMembers(): void {
    if (this.members.length === 0) {
      this.container.innerHTML = `
        <div class="empty-members">
          <p>No family members yet</p>
          <p class="empty-subtitle">Add your first family member to get started</p>
        </div>
      `;
      return;
    }

    this.container.innerHTML = `
      <div class="members-grid">
        ${this.members.map(member => this.renderMemberCard(member)).join('')}
      </div>
    `;
  }

  private renderMemberCard(member: FamilyMember): string {
    const memberTypeClass = `member-type-${member.member_type}`;
    const memberIcon = this.getMemberIcon(member.member_type);

    return `
      <div class="member-card" data-member-id="${member.id}">
        <div class="member-header">
          <div class="member-avatar">
            ${member.avatar_url ? `<img src="${member.avatar_url}" alt="${member.name}" class="member-avatar-img">` : `<span class="member-icon">${memberIcon}</span>`}
          </div>
          <div class="member-actions">
            <button class="member-action-btn" data-action="edit" data-member-id="${member.id}" title="Edit">
              ✎
            </button>
            <button class="member-action-btn" data-action="delete" data-member-id="${member.id}" title="Delete">
              ×
            </button>
          </div>
        </div>
        <div class="member-info">
          <h4 class="member-name">${member.name}</h4>
          ${member.nickname ? `<p class="member-nickname">"${member.nickname}"</p>` : ''}
          ${member.age ? `<p class="member-age">Age: ${member.age}</p>` : ''}
          ${member.user?.email ? `<p class="member-email">${member.user.email}</p>` : ''}
          <span class="member-type ${memberTypeClass}">${member.member_type}</span>
        </div>
      </div>
    `;
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

  private handleClick(e: Event): void {
    const target = e.target as HTMLElement;
    const action = target.getAttribute('data-action');
    const memberId = target.getAttribute('data-member-id');

    if (!action || !memberId) return;

    switch (action) {
      case 'edit':
        this.editMember(memberId);
        break;
      case 'delete':
        this.deleteMember(memberId);
        break;
    }
  }

  private editMember(memberId: string): void {
    const member = this.members.find(m => m.id === memberId);
    if (!member) return;

    this.familyMemberModal.showEdit(member);
  }

  private async handleModalSave(data: FamilyMemberFormData, memberId?: string): Promise<void> {
    const familyId = 'fam1'; // TODO: Get actual family ID

    if (memberId) {
      // Edit existing member
      await this.familyService.updateFamilyMember(familyId, memberId, data);
      this.showSuccess('Member updated successfully!');
    } else {
      // Create new member
      await this.familyService.createFamilyMember(familyId, data);
      this.showSuccess('Member added successfully!');
    }

    await this.loadMembers(); // Refresh the list
  }

  private async deleteMember(memberId: string): Promise<void> {
    const member = this.members.find(m => m.id === memberId);
    if (!member) return;

    const confirmed = confirm(`Are you sure you want to delete ${member.name}?`);
    if (!confirmed) return;

    try {
      const familyId = 'fam1'; // TODO: Get actual family ID
      await this.familyService.deleteFamilyMember(familyId, memberId);
      await this.loadMembers(); // Refresh the list
      this.showSuccess('Member deleted successfully!');
    } catch (error) {
      this.showError('Failed to delete member. Please try again.');
    }
  }

  private showSuccess(message: string): void {
    ComponentUtils.showSuccess(message);
  }

  private showError(message: string): void {
    ComponentUtils.showError(message);
  }

  public async refresh(): Promise<void> {
    await this.loadMembers();
  }

  public destroy(): void {
    if (this.boundHandleClick) {
      this.container.removeEventListener('click', this.boundHandleClick);
    }

    if (this.familyMemberModal) {
      this.familyMemberModal.destroy();
    }
  }
}
