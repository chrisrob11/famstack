import { ComponentConfig } from '../common/types.js';
import { PageComponent } from './base-page.js';
import { DailyPage } from './daily-page.js';
import { FamilyPage } from '../families/family-page.js';
import { SchedulesPage } from '../schedules/schedules-page.js';
import { IntegrationsPage } from './integrations-page.js';

/**
 * Factory for creating page components based on page type
 */
export class PageFactory {
  static createPage(
    pageType: string,
    container: HTMLElement,
    config: ComponentConfig
  ): PageComponent {
    switch (pageType) {
      case 'daily':
        return new DailyPage(container, config);
      case 'family':
        return new FamilyPage(container, config);
      case 'schedules':
        return new SchedulesPage(container, config);
      case 'integrations':
        return new IntegrationsPage(container, config);
      default:
        return new DailyPage(container, config); // Default to daily page
    }
  }

  static getAvailablePageTypes(): string[] {
    return ['daily', 'family', 'schedules', 'integrations'];
  }
}

/**
 * Page manager handles page navigation and lifecycle
 */
export class PageManager {
  private currentPage?: PageComponent;
  private container: HTMLElement;
  private config: ComponentConfig;

  constructor(container: HTMLElement, config: ComponentConfig) {
    this.container = container;
    this.config = config;
  }

  async navigateToPage(pageType: string): Promise<void> {
    // Clean up current page
    if (this.currentPage) {
      this.currentPage.destroy();
    }

    // Create and initialize new page
    this.currentPage = PageFactory.createPage(pageType, this.container, this.config);
    await this.currentPage.init();
  }

  async refreshCurrentPage(): Promise<void> {
    if (this.currentPage && this.currentPage.refresh) {
      await this.currentPage.refresh();
    }
  }

  destroy(): void {
    if (this.currentPage) {
      this.currentPage.destroy();
    }
  }
}
