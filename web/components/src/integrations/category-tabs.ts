/**
 * Category Tabs component
 * Navigation tabs for filtering integrations by category
 */

import { ComponentConfig } from '../common/types.js';
import { INTEGRATION_CATEGORIES } from './integration-types.js';
import { loadCSS } from '../common/dom-utils.js';
import { logger } from '../common/logger.js';

export class CategoryTabs {
  private container: HTMLElement;
  private config: ComponentConfig;
  private activeCategory: string = 'all';

  constructor(container: HTMLElement, config: ComponentConfig) {
    this.container = container;
    this.config = config;
  }

  async init(): Promise<void> {
    await this.loadStyles();
    this.render();
    this.setupEventListeners();
  }

  private async loadStyles(): Promise<void> {
    try {
      await loadCSS(
        '/components/src/integrations/styles/category-tabs.css',
        'category-tabs-styles'
      );
    } catch (error) {
      logger.styleError('CategoryTabs', error);
    }
  }

  private render(): void {
    this.container.innerHTML = `
      <div class="integration-categories">
        <div class="category-tabs">
          <button class="tab-btn active" data-category="all">All</button>
          ${INTEGRATION_CATEGORIES.map(
            category =>
              `<button class="tab-btn" data-category="${category.key}">${category.label}</button>`
          ).join('')}
        </div>
      </div>
    `;
  }

  private setupEventListeners(): void {
    this.container.addEventListener('click', e => {
      const target = e.target as HTMLElement;

      if (target.classList.contains('tab-btn')) {
        const category = target.getAttribute('data-category');
        if (category) {
          this.setActiveCategory(category);
        }
      }
    });

    // Keyboard navigation
    this.container.addEventListener('keydown', e => {
      const target = e.target as HTMLElement;

      if (!target.classList.contains('tab-btn')) return;

      const tabs = Array.from(this.container.querySelectorAll('.tab-btn')) as HTMLElement[];
      const currentIndex = tabs.indexOf(target);

      let nextIndex = currentIndex;

      switch (e.key) {
        case 'ArrowLeft':
          e.preventDefault();
          nextIndex = currentIndex > 0 ? currentIndex - 1 : tabs.length - 1;
          break;
        case 'ArrowRight':
          e.preventDefault();
          nextIndex = currentIndex < tabs.length - 1 ? currentIndex + 1 : 0;
          break;
        case 'Home':
          e.preventDefault();
          nextIndex = 0;
          break;
        case 'End':
          e.preventDefault();
          nextIndex = tabs.length - 1;
          break;
        case 'Enter':
        case ' ':
          e.preventDefault();
          target.click();
          return;
      }

      if (nextIndex !== currentIndex && tabs[nextIndex]) {
        tabs[nextIndex]?.focus();
      }
    });
  }

  private setActiveCategory(category: string): void {
    if (this.activeCategory === category) return;

    this.activeCategory = category;

    // Update visual state
    const tabs = this.container.querySelectorAll('.tab-btn');
    tabs.forEach(tab => {
      const tabCategory = tab.getAttribute('data-category');
      tab.classList.toggle('active', tabCategory === category);
    });

    // Dispatch event
    this.container.dispatchEvent(
      new CustomEvent('category-changed', {
        detail: { category },
        bubbles: true,
      })
    );
  }

  setCategory(category: string): void {
    this.setActiveCategory(category);
  }

  getActiveCategory(): string {
    return this.activeCategory;
  }

  destroy(): void {
    // Component cleanup
  }
}
