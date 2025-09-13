import { ComponentConfig } from '../common/types.js';

export class FamilySelector {
  private container: HTMLElement;
  private selectedUsers: Set<string> = new Set();

  constructor(container: HTMLElement, _config: ComponentConfig) {
    this.container = container;
    this.init();
  }

  private init(): void {
    this.setupEventListeners();
    this.loadInitialSelections();
  }

  private setupEventListeners(): void {
    this.container.addEventListener('click', e => {
      const target = e.target as HTMLElement;

      if (target.matches('.user-avatar') || target.matches('.user-selector-item')) {
        this.toggleUserSelection(target);
      }
    });
  }

  private loadInitialSelections(): void {
    const preselected = this.container.querySelectorAll('.user-selector-item.selected');
    preselected.forEach(item => {
      const userId = item.getAttribute('data-user-id');
      if (userId) {
        this.selectedUsers.add(userId);
      }
    });
  }

  private toggleUserSelection(element: HTMLElement): void {
    const selectorItem = element.closest('.user-selector-item') as HTMLElement;
    if (!selectorItem) return;

    const userId = selectorItem.getAttribute('data-user-id');
    if (!userId) return;

    const isSelected = this.selectedUsers.has(userId);
    if (isSelected) {
      this.selectedUsers.delete(userId);
      selectorItem.classList.remove('selected');
    } else {
      this.selectedUsers.add(userId);
      selectorItem.classList.add('selected');
    }

    this.updateVisualState(selectorItem, !isSelected);
    this.notifySelectionChange();
  }

  private updateVisualState(element: HTMLElement, selected: boolean): void {
    const avatar = element.querySelector('.user-avatar');

    if (selected) {
      avatar?.classList.add('selected');
      element.setAttribute('aria-selected', 'true');
    } else {
      avatar?.classList.remove('selected');
      element.setAttribute('aria-selected', 'false');
    }
  }

  private notifySelectionChange(): void {
    const event = new CustomEvent('familySelectionChange', {
      detail: {
        selectedUsers: Array.from(this.selectedUsers),
      },
    });

    this.container.dispatchEvent(event);

    // Also update any hidden form inputs
    this.updateFormInputs();
  }

  private updateFormInputs(): void {
    const hiddenInputs = this.container.querySelectorAll('input[type="hidden"][name$="[]"]');

    // Remove existing hidden inputs
    hiddenInputs.forEach(input => input.remove());

    // Add new hidden inputs for selected users
    this.selectedUsers.forEach(userId => {
      const input = document.createElement('input');
      input.type = 'hidden';
      input.name = 'assigned_users[]';
      input.value = userId;
      this.container.appendChild(input);
    });
  }

  public getSelectedUsers(): string[] {
    return Array.from(this.selectedUsers);
  }

  public setSelectedUsers(userIds: string[]): void {
    this.selectedUsers.clear();

    userIds.forEach(userId => {
      this.selectedUsers.add(userId);
    });

    // Update visual state
    this.container.querySelectorAll('.user-selector-item').forEach(item => {
      const userId = item.getAttribute('data-user-id');
      const isSelected = userId ? this.selectedUsers.has(userId) : false;

      this.updateVisualState(item as HTMLElement, isSelected);
      item.classList.toggle('selected', isSelected);
    });

    this.updateFormInputs();
  }

  public clearSelection(): void {
    this.selectedUsers.clear();

    this.container.querySelectorAll('.user-selector-item').forEach(item => {
      item.classList.remove('selected');
      this.updateVisualState(item as HTMLElement, false);
    });

    this.updateFormInputs();
  }
}
