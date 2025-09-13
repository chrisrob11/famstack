/**
 * ComponentUtils - Reusable utilities for consistent component patterns
 * Standardizes error handling, loading states, and common UI operations
 */
export class ComponentUtils {
  /**
   * Show loading state in container
   */
  static showLoading(container: HTMLElement, message: string = 'Loading...'): void {
    container.innerHTML = `
      <div class="loading-container">
        <div class="loading-spinner"></div>
        <p>${message}</p>
      </div>
    `;
  }

  /**
   * Show error notification (temporary overlay)
   */
  static showError(message: string, duration: number = 5000): void {
    const errorDiv = document.createElement('div');
    errorDiv.className = 'error-notification';
    errorDiv.style.cssText = `
      position: fixed;
      top: 20px;
      right: 20px;
      background: #f56565;
      color: white;
      padding: 12px 16px;
      border-radius: 4px;
      z-index: 1000;
      max-width: 300px;
      animation: slideIn 0.3s ease;
    `;
    errorDiv.textContent = message;

    document.body.appendChild(errorDiv);

    // Auto-remove after duration
    setTimeout(() => {
      if (errorDiv.parentNode) {
        errorDiv.remove();
      }
    }, duration);
  }

  /**
   * Show error state in container with retry option
   */
  static showErrorState(container: HTMLElement, message: string, onRetry?: () => void): void {
    container.innerHTML = `
      <div class="error-container">
        <div class="error-message">${message}</div>
        ${
          onRetry
            ? `
          <button class="retry-btn" onclick="this.dispatchEvent(new CustomEvent('retry', { bubbles: true }))">
            Try Again
          </button>
        `
            : ''
        }
      </div>
    `;

    if (onRetry) {
      container.addEventListener('retry', onRetry, { once: true });
    }
  }

  /**
   * Show success notification (temporary overlay)
   */
  static showSuccess(message: string, duration: number = 3000): void {
    const successDiv = document.createElement('div');
    successDiv.className = 'success-notification';
    successDiv.style.cssText = `
      position: fixed;
      top: 20px;
      right: 20px;
      background: #48bb78;
      color: white;
      padding: 12px 16px;
      border-radius: 4px;
      z-index: 1000;
      max-width: 300px;
      animation: slideIn 0.3s ease;
    `;
    successDiv.textContent = message;

    document.body.appendChild(successDiv);

    // Auto-remove after duration
    setTimeout(() => {
      if (successDiv.parentNode) {
        successDiv.remove();
      }
    }, duration);
  }

  /**
   * Handle async operations with consistent loading/error patterns
   */
  static async withLoading<T>(
    container: HTMLElement,
    operation: () => Promise<T>,
    loadingMessage?: string
  ): Promise<T> {
    try {
      ComponentUtils.showLoading(container, loadingMessage);
      return await operation();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'An error occurred';
      ComponentUtils.showErrorState(container, errorMessage, () => {
        ComponentUtils.withLoading(container, operation, loadingMessage);
      });
      throw error;
    }
  }

  /**
   * Validate form and show field errors
   */
  static showFormErrors(
    form: HTMLFormElement,
    errors: string | Array<{ field: string; message: string }>
  ): void {
    // Clear existing errors
    form.querySelectorAll('.field-error').forEach(el => el.remove());

    if (typeof errors === 'string') {
      const errorDiv = document.createElement('div');
      errorDiv.className = 'field-error form-error-general';
      errorDiv.textContent = errors;
      form.appendChild(errorDiv);
    } else if (Array.isArray(errors)) {
      errors.forEach(error => {
        const field = form.querySelector(`[name="${error.field}"]`);
        if (field && field.parentNode) {
          const errorDiv = document.createElement('div');
          errorDiv.className = 'field-error';
          errorDiv.textContent = error.message;
          field.parentNode.appendChild(errorDiv);
        }
      });
    }
  }

  /**
   * Clear all form errors
   */
  static clearFormErrors(form: HTMLFormElement): void {
    form.querySelectorAll('.field-error').forEach(el => el.remove());
  }
}
