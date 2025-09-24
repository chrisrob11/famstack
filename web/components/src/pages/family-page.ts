/**
 * Family page component for SPA
 */

import { BasePage } from './base-page.js';
import { ComponentConfig } from '../common/types.js';
import { logger } from '../common/logger.js';

export class FamilyPage extends BasePage {
  constructor(container: HTMLElement, config: ComponentConfig) {
    super(container, config, 'family');
  }

  async init(): Promise<void> {
    this.render();
    this.setupEventListeners();
  }

  private render(): void {
    this.container.innerHTML = `
      <div class="family-page">
        <div class="family-header">
          <h1>Family</h1>
          <p>Manage your family members and settings</p>
        </div>

        <div class="family-content">
          <div class="family-section">
            <h2>Family Members</h2>
            <div id="family-members-list">
              <div class="loading-placeholder">Loading family members...</div>
            </div>
            <button class="btn btn-primary" id="add-member-btn">Add Family Member</button>
          </div>

          <div class="family-section">
            <h2>Family Settings</h2>
            <div class="family-settings">
              <div class="setting-group">
                <label for="family-name">Family Name</label>
                <input type="text" id="family-name" class="form-input" placeholder="Our Family">
              </div>
              <div class="setting-group">
                <label for="family-timezone">Family Timezone</label>
                <select id="family-timezone" class="form-input">
                  <option value="">Loading timezones...</option>
                </select>
              </div>
              <button class="btn btn-secondary" id="save-settings-btn">Save Settings</button>
            </div>
          </div>
        </div>

        <!-- Add Family Member Modal -->
        <div id="add-member-modal" class="modal-overlay" style="display: none;">
          <div class="modal-dialog">
            <div class="modal-header">
              <h3>Add Family Member</h3>
              <button class="modal-close" id="modal-close-btn">&times;</button>
            </div>
            <div class="modal-body">
              <form id="add-member-form">
                <div class="form-group">
                  <label for="member-first-name">First Name</label>
                  <input type="text" id="member-first-name" class="form-input" required>
                </div>
                <div class="form-group">
                  <label for="member-last-name">Last Name</label>
                  <input type="text" id="member-last-name" class="form-input">
                </div>
                <div class="form-group">
                  <label for="member-email">Email Address (optional)</label>
                  <input type="email" id="member-email" class="form-input" placeholder="For future authentication setup">
                  <small style="color: #6b7280; font-size: 0.75rem; margin-top: 0.25rem; display: block;">Email can be added later for login access</small>
                </div>
                <div class="form-group">
                  <label for="member-type">Member Type</label>
                  <select id="member-type" class="form-input">
                    <option value="adult">Adult</option>
                    <option value="child">Child</option>
                    <option value="pet">Pet</option>
                  </select>
                </div>
              </form>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" id="modal-cancel-btn">Cancel</button>
              <button type="submit" class="btn btn-primary" id="modal-save-btn">Add Member</button>
            </div>
          </div>
        </div>
      </div>
    `;

    this.addStyles();
    this.populateTimezones();
    this.loadFamilyData();
  }

  private addStyles(): void {
    const styles = `
      <style id="family-page-styles">
        .family-page {
          padding: 2rem;
          max-width: 1200px;
          margin: 0 auto;
        }

        .family-header {
          margin-bottom: 2rem;
        }

        .family-header h1 {
          font-size: 2rem;
          font-weight: 700;
          color: #374151;
          margin: 0 0 0.5rem 0;
        }

        .family-header p {
          color: #6b7280;
          font-size: 1rem;
          margin: 0;
        }

        .family-content {
          display: grid;
          gap: 2rem;
          grid-template-columns: 1fr 1fr;
        }

        .family-section {
          background: white;
          border-radius: 0.5rem;
          padding: 1.5rem;
          box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
        }

        .family-section h2 {
          font-size: 1.25rem;
          font-weight: 600;
          color: #374151;
          margin: 0 0 1rem 0;
        }

        .family-member {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 0.75rem;
          border: 1px solid #e5e7eb;
          border-radius: 0.375rem;
          margin-bottom: 0.5rem;
        }

        .member-info {
          display: flex;
          align-items: center;
          gap: 0.75rem;
        }

        .member-avatar {
          width: 2rem;
          height: 2rem;
          border-radius: 50%;
          background: #6366f1;
          color: white;
          display: flex;
          align-items: center;
          justify-content: center;
          font-weight: 500;
          font-size: 0.875rem;
        }

        .member-details h3 {
          font-size: 0.875rem;
          font-weight: 500;
          color: #374151;
          margin: 0;
        }

        .member-details p {
          font-size: 0.75rem;
          color: #6b7280;
          margin: 0;
        }

        .setting-group {
          margin-bottom: 1rem;
        }

        .setting-group label {
          display: block;
          margin-bottom: 0.5rem;
          font-weight: 500;
          color: #374151;
          font-size: 0.875rem;
        }

        .form-input {
          width: 100%;
          padding: 0.5rem;
          border: 1px solid #d1d5db;
          border-radius: 0.375rem;
          font-size: 0.875rem;
          box-sizing: border-box;
        }

        .form-input:focus {
          outline: none;
          border-color: #6366f1;
          box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
        }

        .btn {
          padding: 0.5rem 1rem;
          border-radius: 0.375rem;
          font-size: 0.875rem;
          font-weight: 500;
          cursor: pointer;
          border: none;
          transition: background-color 0.2s;
        }

        .btn-primary {
          background: #6366f1;
          color: white;
        }

        .btn-primary:hover {
          background: #5856eb;
        }

        .btn-secondary {
          background: #f3f4f6;
          color: #374151;
        }

        .btn-secondary:hover {
          background: #e5e7eb;
        }

        .loading-placeholder {
          color: #6b7280;
          font-style: italic;
          padding: 1rem;
          text-align: center;
        }

        /* Modal Styles */
        .modal-overlay {
          position: fixed;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background: rgba(0, 0, 0, 0.5);
          display: flex;
          align-items: center;
          justify-content: center;
          z-index: 1000;
        }

        .modal-dialog {
          background: white;
          border-radius: 0.5rem;
          box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
          max-width: 500px;
          width: 90%;
          max-height: 90vh;
          overflow: hidden;
        }

        .modal-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 1.5rem;
          border-bottom: 1px solid #e5e7eb;
        }

        .modal-header h3 {
          font-size: 1.25rem;
          font-weight: 600;
          color: #374151;
          margin: 0;
        }

        .modal-close {
          background: none;
          border: none;
          font-size: 1.5rem;
          color: #6b7280;
          cursor: pointer;
          padding: 0;
          width: 2rem;
          height: 2rem;
          display: flex;
          align-items: center;
          justify-content: center;
          border-radius: 0.25rem;
        }

        .modal-close:hover {
          background: #f3f4f6;
          color: #374151;
        }

        .modal-body {
          padding: 1.5rem;
        }

        .form-group {
          margin-bottom: 1rem;
        }

        .form-group:last-child {
          margin-bottom: 0;
        }

        .form-group label {
          display: block;
          margin-bottom: 0.5rem;
          font-weight: 500;
          color: #374151;
          font-size: 0.875rem;
        }

        .modal-footer {
          display: flex;
          justify-content: flex-end;
          gap: 0.75rem;
          padding: 1.5rem;
          border-top: 1px solid #e5e7eb;
          background: #f9fafb;
        }

        select.form-input {
          background-image: url("data:image/svg+xml;charset=utf-8,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3E%3Cpath stroke='%236b7280' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='m6 8 4 4 4-4'/%3E%3C/svg%3E");
          background-position: right 0.5rem center;
          background-repeat: no-repeat;
          background-size: 1.5rem 1.5rem;
          padding-right: 2.5rem;
          appearance: none;
        }

        @media (max-width: 768px) {
          .family-content {
            grid-template-columns: 1fr;
          }

          .modal-dialog {
            width: 95%;
            margin: 1rem;
          }

          .modal-footer {
            flex-direction: column-reverse;
          }

          .modal-footer .btn {
            width: 100%;
          }
        }
      </style>
    `;

    // Remove existing styles
    const existingStyles = document.getElementById('family-page-styles');
    if (existingStyles) {
      existingStyles.remove();
    }

    // Add styles to head
    document.head.insertAdjacentHTML('beforeend', styles);
  }

  private setupEventListeners(): void {
    const addMemberBtn = document.getElementById('add-member-btn');
    const saveSettingsBtn = document.getElementById('save-settings-btn');

    // Modal elements
    const modal = document.getElementById('add-member-modal');
    const modalCloseBtn = document.getElementById('modal-close-btn');
    const modalCancelBtn = document.getElementById('modal-cancel-btn');
    const modalSaveBtn = document.getElementById('modal-save-btn');
    const addMemberForm = document.getElementById('add-member-form');

    if (addMemberBtn) {
      addMemberBtn.addEventListener('click', () => this.showAddMemberModal());
    }

    if (saveSettingsBtn) {
      saveSettingsBtn.addEventListener('click', () => this.handleSaveSettings());
    }

    // Modal event listeners
    if (modalCloseBtn) {
      modalCloseBtn.addEventListener('click', () => this.hideAddMemberModal());
    }

    if (modalCancelBtn) {
      modalCancelBtn.addEventListener('click', () => this.hideAddMemberModal());
    }

    if (modalSaveBtn) {
      modalSaveBtn.addEventListener('click', e => {
        e.preventDefault();
        this.handleModalSave();
      });
    }

    if (addMemberForm) {
      addMemberForm.addEventListener('submit', e => {
        e.preventDefault();
        this.handleModalSave();
      });
    }

    // Close modal when clicking on overlay
    if (modal) {
      modal.addEventListener('click', e => {
        if (e.target === modal) {
          this.hideAddMemberModal();
        }
      });
    }

    // Close modal with Escape key
    document.addEventListener('keydown', e => {
      if (e.key === 'Escape') {
        this.hideAddMemberModal();
      }
    });
  }

  private async loadFamilyData(): Promise<void> {
    try {
      // Get current user session to extract family ID
      const authResponse = await fetch('/auth/me', {
        credentials: 'include',
      });
      if (!authResponse.ok) {
        logger.warn('Failed to get user session');
        this.renderFamilyMembers([]);
        return;
      }

      const sessionData = await authResponse.json();
      const familyId = sessionData.session?.family_id || sessionData.user?.family_id;

      if (!familyId) {
        logger.warn('No family ID found in session');
        this.renderFamilyMembers([]);
        return;
      }

      // Load family members using hierarchical API
      const response = await fetch(`/api/v1/families/${familyId}/members`, {
        credentials: 'include',
      });
      if (response.ok) {
        const contentType = response.headers.get('content-type');
        if (contentType && contentType.includes('application/json')) {
          const data = await response.json();
          // The API returns data wrapped in family_members
          const members = data.family_members || [];
          this.renderFamilyMembers(members);
        } else {
          logger.warn('Family members API returned non-JSON response');
          this.renderFamilyMembers([]);
        }
      } else {
        logger.warn(`Family members API failed with status: ${response.status}`);
        this.renderFamilyMembers([]);
      }

      // Load family settings using family ID from session
      try {
        const settingsResponse = await fetch(`/api/v1/families/${familyId}`, {
          credentials: 'include',
        });
        if (settingsResponse.ok) {
          const contentType = settingsResponse.headers.get('content-type');
          if (contentType && contentType.includes('application/json')) {
            const settings = await settingsResponse.json();
            this.populateSettings(settings);
          }
        }
      } catch (settingsError) {
        logger.info('Family settings API not available yet');
      }
    } catch (error) {
      logger.error('Failed to load family data:', error);
      this.renderFamilyMembers([]);
    }
  }

  private renderFamilyMembers(members: any[]): void {
    const container = document.getElementById('family-members-list');
    if (!container) return;

    if (members.length === 0) {
      container.innerHTML = `
        <div class="loading-placeholder">
          No family members found. Add your first member to get started.
        </div>
      `;
      return;
    }

    container.innerHTML = members
      .map(member => {
        // Construct full name from first_name and last_name
        const fullName =
          member.first_name && member.last_name
            ? `${member.first_name} ${member.last_name}`.trim()
            : member.first_name || member.last_name || 'Unknown';

        return `
      <div class="family-member">
        <div class="member-info">
          <div class="member-avatar">
            ${fullName.charAt(0).toUpperCase()}
          </div>
          <div class="member-details">
            <h3>${fullName}</h3>
            <p>${member.member_type || member.role || 'Member'}</p>
          </div>
        </div>
        <button class="btn btn-secondary" onclick="this.closest('.family-page').dispatchEvent(new CustomEvent('edit-member', { detail: { id: '${member.id}' } }))">
          Edit
        </button>
      </div>
    `;
      })
      .join('');
  }

  private populateSettings(settings: any): void {
    const familyNameInput = document.getElementById('family-name') as HTMLInputElement;
    const familyTimezoneSelect = document.getElementById('family-timezone') as HTMLSelectElement;

    if (familyNameInput && settings.name) {
      familyNameInput.value = settings.name;
    }

    if (familyTimezoneSelect && settings.timezone) {
      familyTimezoneSelect.value = settings.timezone;
    }
  }

  private populateTimezones(): void {
    const timezoneSelect = document.getElementById('family-timezone') as HTMLSelectElement;
    if (!timezoneSelect) return;

    // Common timezones organized by regions
    const timezones = [
      { group: 'US & Canada', timezones: [
        { value: 'America/New_York', label: 'Eastern Time (New York)' },
        { value: 'America/Chicago', label: 'Central Time (Chicago)' },
        { value: 'America/Denver', label: 'Mountain Time (Denver)' },
        { value: 'America/Los_Angeles', label: 'Pacific Time (Los Angeles)' },
        { value: 'America/Anchorage', label: 'Alaska Time (Anchorage)' },
        { value: 'Pacific/Honolulu', label: 'Hawaii Time (Honolulu)' },
        { value: 'America/Toronto', label: 'Eastern Time (Toronto)' },
        { value: 'America/Vancouver', label: 'Pacific Time (Vancouver)' }
      ]},
      { group: 'Europe', timezones: [
        { value: 'Europe/London', label: 'London, UK' },
        { value: 'Europe/Paris', label: 'Paris, France' },
        { value: 'Europe/Berlin', label: 'Berlin, Germany' },
        { value: 'Europe/Rome', label: 'Rome, Italy' },
        { value: 'Europe/Madrid', label: 'Madrid, Spain' },
        { value: 'Europe/Amsterdam', label: 'Amsterdam, Netherlands' },
        { value: 'Europe/Stockholm', label: 'Stockholm, Sweden' },
        { value: 'Europe/Moscow', label: 'Moscow, Russia' }
      ]},
      { group: 'Asia Pacific', timezones: [
        { value: 'Asia/Tokyo', label: 'Tokyo, Japan' },
        { value: 'Asia/Shanghai', label: 'Beijing, China' },
        { value: 'Asia/Hong_Kong', label: 'Hong Kong' },
        { value: 'Asia/Singapore', label: 'Singapore' },
        { value: 'Asia/Seoul', label: 'Seoul, South Korea' },
        { value: 'Asia/Kolkata', label: 'Mumbai, India' },
        { value: 'Australia/Sydney', label: 'Sydney, Australia' },
        { value: 'Australia/Melbourne', label: 'Melbourne, Australia' },
        { value: 'Pacific/Auckland', label: 'Auckland, New Zealand' }
      ]},
      { group: 'Others', timezones: [
        { value: 'UTC', label: 'UTC (Coordinated Universal Time)' },
        { value: 'America/Sao_Paulo', label: 'São Paulo, Brazil' },
        { value: 'America/Mexico_City', label: 'Mexico City, Mexico' },
        { value: 'Africa/Johannesburg', label: 'Johannesburg, South Africa' },
        { value: 'Africa/Cairo', label: 'Cairo, Egypt' },
        { value: 'Asia/Dubai', label: 'Dubai, UAE' }
      ]}
    ];

    // Clear existing options
    timezoneSelect.innerHTML = '';

    // Add placeholder option
    const placeholderOption = document.createElement('option');
    placeholderOption.value = '';
    placeholderOption.textContent = 'Select timezone...';
    timezoneSelect.appendChild(placeholderOption);

    // Add timezone options grouped by region
    timezones.forEach(group => {
      const optgroup = document.createElement('optgroup');
      optgroup.label = group.group;

      group.timezones.forEach(tz => {
        const option = document.createElement('option');
        option.value = tz.value;
        option.textContent = tz.label;
        optgroup.appendChild(option);
      });

      timezoneSelect.appendChild(optgroup);
    });
  }

  private showAddMemberModal(): void {
    const modal = document.getElementById('add-member-modal');
    if (modal) {
      modal.style.display = 'flex';
      // Focus on first input
      const firstNameInput = document.getElementById('member-first-name') as HTMLInputElement;
      if (firstNameInput) {
        setTimeout(() => firstNameInput.focus(), 100);
      }
    }
  }

  private hideAddMemberModal(): void {
    const modal = document.getElementById('add-member-modal');
    if (modal) {
      modal.style.display = 'none';
      // Clear form
      this.clearAddMemberForm();
    }
  }

  private clearAddMemberForm(): void {
    const firstNameInput = document.getElementById('member-first-name') as HTMLInputElement;
    const lastNameInput = document.getElementById('member-last-name') as HTMLInputElement;
    const emailInput = document.getElementById('member-email') as HTMLInputElement;
    const memberTypeSelect = document.getElementById('member-type') as HTMLSelectElement;

    if (firstNameInput) firstNameInput.value = '';
    if (lastNameInput) lastNameInput.value = '';
    if (emailInput) emailInput.value = '';
    if (memberTypeSelect) memberTypeSelect.value = 'adult';
  }

  private async handleModalSave(): Promise<void> {
    const firstNameInput = document.getElementById('member-first-name') as HTMLInputElement;
    const lastNameInput = document.getElementById('member-last-name') as HTMLInputElement;
    const emailInput = document.getElementById('member-email') as HTMLInputElement;
    const memberTypeSelect = document.getElementById('member-type') as HTMLSelectElement;

    const firstName = firstNameInput?.value?.trim() || '';
    const lastName = lastNameInput?.value?.trim() || '';
    const email = emailInput?.value?.trim() || '';
    const memberType = memberTypeSelect?.value || 'adult';

    if (!firstName) {
      alert('Please enter a first name');
      if (firstNameInput) firstNameInput.focus();
      return;
    }

    // Email validation only if provided
    if (email) {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(email)) {
        alert('Please enter a valid email address');
        if (emailInput) emailInput.focus();
        return;
      }
    }

    try {
      await this.addFamilyMember(firstName, lastName, memberType);
      this.hideAddMemberModal();
    } catch (error) {
      // Error is already handled in addFamilyMember
    }
  }

  private async addFamilyMember(
    firstName: string,
    lastName: string,
    memberType: string = 'adult'
  ): Promise<void> {
    try {
      // Get family ID from session
      const authResponse = await fetch('/auth/me', {
        credentials: 'include',
      });
      if (!authResponse.ok) {
        throw new Error('Failed to get user session');
      }

      const sessionData = await authResponse.json();
      const familyId = sessionData.session?.family_id || sessionData.user?.family_id;

      if (!familyId) {
        throw new Error('No family ID found');
      }

      const response = await fetch(`/api/v1/families/${familyId}/members`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          first_name: firstName,
          last_name: lastName,
          member_type: memberType,
        }),
      });

      if (response.ok) {
        await this.loadFamilyData(); // Refresh the list
        // Show success message
        const fullName = lastName ? `${firstName} ${lastName}` : firstName;
        alert(`${fullName} has been added to your family!`);
      } else {
        const errorData = await response.text();
        throw new Error(`Failed to add family member: ${response.status} ${errorData}`);
      }
    } catch (error) {
      logger.error('Failed to add family member:', error);
      alert(
        `Failed to add family member: ${error instanceof Error ? error.message : 'Unknown error'}`
      );
      throw error; // Re-throw so modal doesn't close on error
    }
  }

  private async handleSaveSettings(): Promise<void> {
    const familyNameInput = document.getElementById('family-name') as HTMLInputElement;
    const familyTimezoneSelect = document.getElementById('family-timezone') as HTMLSelectElement;

    const familyName = familyNameInput?.value?.trim();
    const familyTimezone = familyTimezoneSelect?.value?.trim();

    if (!familyName) {
      alert('Please enter a family name');
      return;
    }

    if (!familyTimezone) {
      alert('Please select a family timezone');
      return;
    }

    try {
      // Get family ID from session (same as we do in loadFamilyData)
      const authResponse = await fetch('/auth/me', {
        credentials: 'include',
      });
      if (!authResponse.ok) {
        alert('Failed to get user session');
        return;
      }

      const sessionData = await authResponse.json();
      const familyId = sessionData.session?.family_id || sessionData.user?.family_id;

      if (!familyId) {
        alert('Family ID not found in session');
        return;
      }

      // Update family using the family ID from session
      const response = await fetch(`/api/v1/families/${familyId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          name: familyName,
          timezone: familyTimezone
        }),
      });

      if (response.ok) {
        alert('Family settings saved successfully');
      } else {
        alert('Failed to save family settings. Please check that the timezone is valid.');
      }
    } catch (error) {
      logger.error('Failed to save family settings:', error);
      alert('Failed to save family settings');
    }
  }

  async refresh(): Promise<void> {
    await this.loadFamilyData();
  }

  override destroy(): void {
    // Remove styles
    const styles = document.getElementById('family-page-styles');
    if (styles) {
      styles.remove();
    }

    super.destroy();
  }
}

export default FamilyPage;
