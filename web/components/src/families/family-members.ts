import { ComponentConfig } from '../common/types.js';

export interface FamilyMember {
  id: string;
  family_id: string;
  name: string;
  email: string;
  role: string;
  created_at: string;
}

export class FamilyMembers {
  private config: ComponentConfig;
  private container: HTMLElement;
  private members: FamilyMember[] = [];

  constructor(container: HTMLElement, config: ComponentConfig) {
    this.container = container;
    this.config = config;
    this.init();
  }

  private async init(): Promise<void> {
    this.container.className = 'family-members-container';
    
    // Listen for refresh events
    this.container.addEventListener('refresh', () => {
      this.loadMembers();
    });
    
    await this.loadMembers();
  }

  private async loadMembers(): Promise<void> {
    try {
      this.showLoading();
      const response = await fetch(`${this.config.apiBaseUrl}/users`);
      
      if (!response.ok) {
        throw new Error(`Failed to load members: ${response.statusText}`);
      }
      
      this.members = await response.json();
      this.renderMembers();
    } catch (error) {
      this.showError('Failed to load family members');
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

  private showError(message: string): void {
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
    const roleColor = member.role === 'parent' ? 'role-parent' : 'role-child';
    const roleIcon = member.role === 'parent' ? '👨‍👩‍👧‍👦' : '👶';
    
    return `
      <div class="member-card" data-member-id="${member.id}">
        <div class="member-header">
          <div class="member-avatar">
            <span class="member-icon">${roleIcon}</span>
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
          ${member.email ? `<p class="member-email">${member.email}</p>` : ''}
          <span class="member-role ${roleColor}">${member.role}</span>
        </div>
      </div>
    `;
  }

  public async refresh(): Promise<void> {
    await this.loadMembers();
  }

  public destroy(): void {
    // Clean up any event listeners if needed
  }
}