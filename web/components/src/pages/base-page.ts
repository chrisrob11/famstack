import { ComponentConfig } from '../common/types.js';

/**
 * Base interface for all page components
 */
export interface PageComponent {
  /**
   * Initialize the page component
   */
  init(): Promise<void>;
  
  /**
   * Clean up the page component when navigating away
   */
  destroy(): void;
  
  /**
   * Refresh the page component data
   */
  refresh?(): Promise<void>;
}

/**
 * Abstract base class for page components
 */
export abstract class BasePage implements PageComponent {
  protected container: HTMLElement;
  protected config: ComponentConfig;
  protected pageType: string;

  constructor(container: HTMLElement, config: ComponentConfig, pageType: string) {
    this.container = container;
    this.config = config;
    this.pageType = pageType;
  }

  /**
   * Initialize the page - must be implemented by subclasses
   */
  abstract init(): Promise<void>;

  /**
   * Clean up the page - can be overridden by subclasses
   */
  destroy(): void {
    // Default cleanup - remove all event listeners and clear container
    this.container.innerHTML = '';
  }

  /**
   * Show loading state
   */
  protected showLoading(message: string = 'Loading...'): void {
    this.container.innerHTML = `
      <div class="page-loading">
        <div class="loading-spinner"></div>
        <p>${message}</p>
      </div>
    `;
  }

  /**
   * Show error state
   */
  protected showError(message: string, canRetry: boolean = true): void {
    this.container.innerHTML = `
      <div class="page-error">
        <h3>Something went wrong</h3>
        <p>${message}</p>
        ${canRetry ? `
          <button class="btn btn-primary" onclick="this.closest('.page-error').dispatchEvent(new CustomEvent('retry'))">
            Try Again
          </button>
        ` : ''}
      </div>
    `;

    if (canRetry) {
      this.container.addEventListener('retry', () => this.init());
    }
  }

  /**
   * Utility method to create DOM elements
   */
  protected createElement(tag: string, className?: string, innerHTML?: string): HTMLElement {
    const element = document.createElement(tag);
    if (className) element.className = className;
    if (innerHTML) element.innerHTML = innerHTML;
    return element;
  }
}