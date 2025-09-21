/**
 * Login page component for SPA
 */

import { BasePage } from './base-page.js';
import { ComponentConfig } from '../common/types.js';
import { logger } from '../common/logger.js';

export class LoginPage extends BasePage {
  private authManager: any;

  constructor(container: HTMLElement, config: ComponentConfig) {
    super(container, config, 'login');
    this.authManager = (window as any).famstackApp?.getAuthManager();
  }

  async init(): Promise<void> {
    this.render();
    this.setupEventListeners();
  }

  private render(): void {
    this.container.innerHTML = `
      <div class="login-container">
        <div class="login-card">
          <div class="login-header">
            <h1 class="login-title">FamStack</h1>
            <p class="login-subtitle">Family task management made simple</p>
          </div>

          <div id="message-container"></div>

          <form id="login-form">
            <div class="form-group">
              <label for="email" class="form-label">Email</label>
              <input type="email" id="email" name="email" class="form-input" required autocomplete="email">
            </div>
            <div class="form-group">
              <label for="password" class="form-label">Password</label>
              <input type="password" id="password" name="password" class="form-input" required autocomplete="current-password">
            </div>
            <button type="submit" class="login-button" id="login-submit">Sign In</button>
          </form>

          <div class="login-footer">
            <p class="login-help">
              Need help? Contact your family administrator.
            </p>
          </div>
        </div>
      </div>
    `;

    this.addStyles();
  }

  private addStyles(): void {
    const styles = `
      <style id="login-page-styles">
        .login-container {
          min-height: 100vh;
          display: flex;
          align-items: center;
          justify-content: center;
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          padding: 1rem;
        }

        .login-card {
          background: white;
          border-radius: 1rem;
          box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
          padding: 3rem;
          width: 100%;
          max-width: 400px;
        }

        .login-header {
          text-align: center;
          margin-bottom: 2rem;
        }

        .login-title {
          font-size: 2rem;
          font-weight: 700;
          color: #374151;
          margin: 0 0 0.5rem 0;
        }

        .login-subtitle {
          color: #6b7280;
          font-size: 0.875rem;
          margin: 0;
        }

        .form-group {
          margin-bottom: 1.5rem;
        }

        .form-label {
          display: block;
          margin-bottom: 0.5rem;
          font-weight: 500;
          color: #374151;
        }

        .form-input {
          width: 100%;
          padding: 0.75rem;
          border: 1px solid #d1d5db;
          border-radius: 0.5rem;
          font-size: 1rem;
          transition: border-color 0.2s;
          box-sizing: border-box;
        }

        .form-input:focus {
          outline: none;
          border-color: #6366f1;
          box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
        }

        .login-button {
          width: 100%;
          background: #6366f1;
          color: white;
          border: none;
          padding: 0.75rem;
          border-radius: 0.5rem;
          font-size: 1rem;
          font-weight: 500;
          cursor: pointer;
          transition: background-color 0.2s;
          margin-bottom: 1rem;
        }

        .login-button:hover:not(:disabled) {
          background: #5856eb;
        }

        .login-button:disabled {
          background: #9ca3af;
          cursor: not-allowed;
        }

        .login-footer {
          text-align: center;
          margin-top: 1.5rem;
          padding-top: 1.5rem;
          border-top: 1px solid #e5e7eb;
        }

        .login-help {
          color: #6b7280;
          font-size: 0.875rem;
          margin: 0;
        }

        .error-message {
          background: #fef2f2;
          border: 1px solid #fecaca;
          color: #dc2626;
          padding: 0.75rem;
          border-radius: 0.5rem;
          margin-bottom: 1rem;
          font-size: 0.875rem;
        }

        .success-message {
          background: #f0fdf4;
          border: 1px solid #bbf7d0;
          color: #166534;
          padding: 0.75rem;
          border-radius: 0.5rem;
          margin-bottom: 1rem;
          font-size: 0.875rem;
        }

        .loading {
          opacity: 0.6;
          pointer-events: none;
        }

        @media (max-width: 480px) {
          .login-card {
            padding: 2rem;
          }

          .login-title {
            font-size: 1.75rem;
          }
        }
      </style>
    `;

    // Remove existing styles
    const existingStyles = document.getElementById('login-page-styles');
    if (existingStyles) {
      existingStyles.remove();
    }

    // Add styles to head
    document.head.insertAdjacentHTML('beforeend', styles);
  }

  private setupEventListeners(): void {
    const form = document.getElementById('login-form') as HTMLFormElement;
    if (form) {
      form.addEventListener('submit', e => this.handleLogin(e));
    }

    // Auto-focus email input
    const emailInput = document.getElementById('email') as HTMLInputElement;
    if (emailInput) {
      emailInput.focus();
    }
  }

  private async handleLogin(event: Event): Promise<void> {
    event.preventDefault();

    const form = event.target as HTMLFormElement;
    const formData = new FormData(form);
    const email = formData.get('email') as string;
    const password = formData.get('password') as string;

    if (!email || !password) {
      this.showError('Please enter both email and password');
      return;
    }

    this.setLoading(true);
    this.clearMessages();

    try {
      const success = await this.authManager.login({ email, password });

      if (success) {
        this.showSuccess('Login successful! Redirecting...');

        // Wait a moment then redirect
        setTimeout(() => {
          const appContainer = document.getElementById('app');
          const router = (window as any).famstackApp?.getRouter();

          logger.debug('Login redirect debug', {
            appContainerExists: !!appContainer,
            routerExists: !!router,
            currentLocation: window.location.pathname,
          });

          if (!appContainer) {
            logger.error('App container not found during redirect');
            return;
          }

          if (router) {
            logger.debug('Calling router.navigate("/tasks")');
            try {
              router.navigate('/tasks');
              logger.debug('router.navigate() completed');
            } catch (error) {
              logger.error('Error during navigation:', error);
            }
          } else {
            logger.error('Router not found during redirect');
          }
        }, 500);
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Login failed';
      this.showError(message);
    } finally {
      this.setLoading(false);
    }
  }

  protected override showError(message: string): void {
    this.showMessage(message, 'error');
  }

  private showSuccess(message: string): void {
    this.showMessage(message, 'success');
  }

  private showMessage(message: string, type: 'error' | 'success'): void {
    const container = document.getElementById('message-container');
    if (container) {
      const messageEl = document.createElement('div');
      messageEl.className = type === 'error' ? 'error-message' : 'success-message';
      messageEl.textContent = message;
      container.innerHTML = '';
      container.appendChild(messageEl);
    }
  }

  private clearMessages(): void {
    const container = document.getElementById('message-container');
    if (container) {
      container.innerHTML = '';
    }
  }

  private setLoading(loading: boolean): void {
    const button = document.getElementById('login-submit') as HTMLButtonElement;
    const card = document.querySelector('.login-card');

    if (button) {
      button.disabled = loading;
      button.textContent = loading ? 'Signing In...' : 'Sign In';
    }

    if (card) {
      card.classList.toggle('loading', loading);
    }
  }

  override destroy(): void {
    super.destroy();

    // Remove styles
    const styles = document.getElementById('login-page-styles');
    if (styles) {
      styles.remove();
    }
  }
}

export default LoginPage;
