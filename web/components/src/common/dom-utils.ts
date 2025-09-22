/**
 * DOM Utility Functions
 * Shared utilities for safer DOM manipulation
 */

import sanitizeHtml from 'sanitize-html';

/**
 * Type-safe element selection with better error handling
 */
export function getElement<T extends HTMLElement = HTMLElement>(
  selector: string,
  context: Document | HTMLElement = document
): T | null {
  return context.querySelector(selector) as T | null;
}

/**
 * Get element by ID with type safety
 */
export function getElementById<T extends HTMLElement = HTMLElement>(id: string): T | null {
  return document.getElementById(id) as T | null;
}

/**
 * Get required element (throws if not found)
 */
export function getRequiredElement<T extends HTMLElement = HTMLElement>(
  selector: string,
  context: Document | HTMLElement = document
): T {
  const element = getElement<T>(selector, context);
  if (!element) {
    throw new Error(`Required element not found: ${selector}`);
  }
  return element;
}

/**
 * Get required element by ID (throws if not found)
 */
export function getRequiredElementById<T extends HTMLElement = HTMLElement>(id: string): T {
  const element = getElementById<T>(id);
  if (!element) {
    throw new Error(`Required element not found with ID: ${id}`);
  }
  return element;
}

/**
 * Get all elements matching selector
 */
export function getElements<T extends HTMLElement = HTMLElement>(
  selector: string,
  context: Document | HTMLElement = document
): T[] {
  return Array.from(context.querySelectorAll(selector)) as T[];
}

/**
 * Safely set innerHTML with proper XSS protection
 */
export function setInnerHTML(element: HTMLElement, html: string): void {
  // Use proper HTML sanitization library to prevent XSS
  const sanitized = sanitizeHtml(html, {
    allowedTags: [
      'b',
      'i',
      'em',
      'strong',
      'a',
      'p',
      'br',
      'ul',
      'ol',
      'li',
      'div',
      'span',
      'h1',
      'h2',
      'h3',
      'h4',
      'h5',
      'h6',
      'blockquote',
      'code',
      'pre',
    ],
    allowedAttributes: {
      a: ['href', 'title'],
      div: ['class'],
      span: ['class'],
      p: ['class'],
      h1: ['class'],
      h2: ['class'],
      h3: ['class'],
      h4: ['class'],
      h5: ['class'],
      h6: ['class'],
    },
    allowedSchemes: ['http', 'https', 'mailto'],
    disallowedTagsMode: 'discard',
    allowedSchemesByTag: {},
    allowedSchemesAppliedToAttributes: ['href'],
  });

  element.innerHTML = sanitized;
}

/**
 * Create element with attributes and content
 */
export function createElement<K extends keyof HTMLElementTagNameMap>(
  tagName: K,
  options: {
    className?: string;
    id?: string;
    textContent?: string;
    innerHTML?: string;
    attributes?: Record<string, string>;
  } = {}
): HTMLElementTagNameMap[K] {
  const element = document.createElement(tagName);

  if (options.className) element.className = options.className;
  if (options.id) element.id = options.id;
  if (options.textContent) element.textContent = options.textContent;
  if (options.innerHTML) setInnerHTML(element, options.innerHTML);

  if (options.attributes) {
    Object.entries(options.attributes).forEach(([key, value]) => {
      element.setAttribute(key, value);
    });
  }

  return element;
}

/**
 * Add event listener with automatic cleanup
 */
export function addEventListenerWithCleanup<K extends keyof HTMLElementEventMap>(
  element: HTMLElement,
  type: K,
  listener: (this: HTMLElement, ev: HTMLElementEventMap[K]) => any,
  options?: boolean | AddEventListenerOptions
): () => void {
  element.addEventListener(type, listener, options);

  return () => {
    element.removeEventListener(type, listener, options);
  };
}

/**
 * Toggle element visibility
 */
export function toggleVisibility(element: HTMLElement, visible?: boolean): void {
  if (visible === undefined) {
    visible = element.style.display === 'none';
  }
  element.style.display = visible ? '' : 'none';
}

/**
 * Load CSS file dynamically
 */
export function loadCSS(href: string, id?: string): Promise<void> {
  return new Promise((resolve, reject) => {
    // Check if already loaded
    if (id && document.getElementById(id)) {
      resolve();
      return;
    }

    const link = document.createElement('link');
    link.rel = 'stylesheet';
    link.href = href;
    if (id) link.id = id;

    link.onload = () => resolve();
    link.onerror = () => reject(new Error(`Failed to load CSS: ${href}`));

    document.head.appendChild(link);
  });
}

/**
 * Debounce function for event handlers
 */
export function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: NodeJS.Timeout;

  return (...args: Parameters<T>) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), wait);
  };
}

/**
 * Form data extraction with type safety
 */
export function getFormData<T extends Record<string, any>>(form: HTMLFormElement): Partial<T> {
  const formData = new FormData(form);
  const result: Record<string, any> = {};

  for (const [key, value] of formData.entries()) {
    if (typeof value === 'string') {
      result[key] = value;
    }
  }

  return result as Partial<T>;
}

/**
 * Validate form and show validation messages
 */
export function validateForm(form: HTMLFormElement): boolean {
  const isValid = form.checkValidity();
  if (!isValid) {
    form.reportValidity();
  }
  return isValid;
}

/**
 * Escape HTML to prevent XSS
 */
export function escapeHtml(text: string): string {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

/**
 * Generate unique ID
 */
export function generateId(prefix = 'id'): string {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).substring(2, 11)}`;
}
