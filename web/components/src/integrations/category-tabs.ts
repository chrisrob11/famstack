/**
 * CategoryTabs Component
 *
 * Role: Navigation tabs for filtering integrations by category
 * Responsibilities:
 * - Display integration category tabs (All, Calendar, Communication, etc.)
 * - Handle tab selection and active state management
 * - Provide keyboard navigation support (arrow keys, home/end)
 * - Dispatch category change events
 * - Maintain accessibility with proper ARIA attributes
 */

import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { INTEGRATION_CATEGORIES } from './integration-types.js';
import { EVENTS } from '../common/constants.js';

@customElement('category-tabs')
export class CategoryTabs extends LitElement {
  @property({ type: String, attribute: 'active-category' })
  activeCategory = 'all';

  static override styles = css`
    :host {
      display: block;
    }

    .integration-categories {
      margin-bottom: 20px;
    }

    .category-tabs {
      display: flex;
      gap: 4px;
      border-bottom: 2px solid #e1e5e9;
      overflow-x: auto;
      scrollbar-width: thin;
    }

    .tab-btn {
      background: none;
      border: none;
      padding: 12px 20px;
      font-size: 14px;
      font-weight: 500;
      color: #6c757d;
      cursor: pointer;
      border-bottom: 2px solid transparent;
      transition: all 0.2s ease;
      white-space: nowrap;
      position: relative;
      top: 2px;
    }

    .tab-btn:hover {
      color: #495057;
      background: #f8f9fa;
    }

    .tab-btn:focus {
      outline: 2px solid #007bff;
      outline-offset: -2px;
    }

    .tab-btn.active {
      color: #007bff;
      border-bottom-color: #007bff;
      background: #fff;
    }

    .tab-btn:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }
  `;

  private setActiveCategory(category: string): void {
    if (this.activeCategory === category) return;

    this.activeCategory = category;

    this.dispatchEvent(
      new CustomEvent(EVENTS.CATEGORY_CHANGED, {
        detail: { category },
        bubbles: true,
      })
    );
  }

  private handleTabClick(e: Event) {
    const target = e.target as HTMLElement;
    const category = target.getAttribute('data-category');
    if (category) {
      this.setActiveCategory(category);
    }
  }

  private handleKeyDown(e: KeyboardEvent) {
    const target = e.target as HTMLElement;
    if (!target.classList.contains('tab-btn')) return;

    const tabs = Array.from(this.shadowRoot?.querySelectorAll('.tab-btn') || []) as HTMLElement[];
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
  }

  setCategory(category: string): void {
    this.setActiveCategory(category);
  }

  getActiveCategory(): string {
    return this.activeCategory;
  }

  override render() {
    return html`
      <div class="integration-categories">
        <div class="category-tabs" @click=${this.handleTabClick} @keydown=${this.handleKeyDown}>
          <button
            class="tab-btn ${this.activeCategory === 'all' ? 'active' : ''}"
            data-category="all"
            tabindex="${this.activeCategory === 'all' ? '0' : '-1'}"
          >
            All
          </button>
          ${INTEGRATION_CATEGORIES.map(
            category => html`
              <button
                class="tab-btn ${this.activeCategory === category.key ? 'active' : ''}"
                data-category="${category.key}"
                tabindex="${this.activeCategory === category.key ? '0' : '-1'}"
              >
                ${category.label}
              </button>
            `
          )}
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'category-tabs': CategoryTabs;
  }
}
