/**
 * FamilyPage Wrapper - Bridges between Lit component and page factory
 */

import { BasePage } from '../pages/base-page.js';
import { ComponentConfig } from '../common/types.js';
import './family-page.js';

export class FamilyPageWrapper extends BasePage {
  private litElement: HTMLElement | undefined = undefined;

  constructor(container: HTMLElement, config: ComponentConfig) {
    super(container, config, 'family');
  }

  async init(): Promise<void> {
    this.render();
  }

  private render(): void {
    // Clear container and add the Lit component
    this.container.innerHTML = '';
    this.litElement = document.createElement('family-page');
    this.container.appendChild(this.litElement);
  }

  async refresh(): Promise<void> {
    // The Lit component handles its own refresh
    // We could dispatch an event if needed
  }

  override destroy(): void {
    if (this.litElement) {
      this.litElement.remove();
    }
    this.litElement = undefined;
    super.destroy();
  }
}

export default FamilyPageWrapper;
