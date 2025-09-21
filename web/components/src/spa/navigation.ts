/**
 * Navigation component for SPA
 * Unified navigation that works across all pages
 */

import { AuthManager } from './auth-manager.js';
import { SPARouter } from './router.js';
import { logger } from '../common/logger.js';

export interface NavigationConfig {
  authManager: AuthManager;
  router: SPARouter;
}

export class NavigationComponent {
  private container: HTMLElement;
  private authManager: AuthManager;
  private router: SPARouter;
  private isDropdownOpen = false;

  constructor(container: HTMLElement, config: NavigationConfig) {
    this.container = container;
    this.authManager = config.authManager;
    this.router = config.router;
  }

  async init(): Promise<void> {
    this.render();
    this.setupEventListeners();
    this.updateAuthState();
  }

  private render(): void {
    this.container.innerHTML = `
      <nav class="main-nav">
        <div class="nav-container">
          <h1 class="nav-title">
            <a href="/" data-route="/">FamStack</a>
          </h1>
          <div class="nav-menu">
            <div class="dropdown">
              <button class="dropdown-btn" aria-expanded="false" id="nav-dropdown-btn">
                Menu ▼
              </button>
              <div class="dropdown-content" id="nav-dropdown-content">
                <a href="/tasks" data-route="/tasks">Daily Tasks</a>
                <a href="/schedules" data-route="/schedules">Schedules</a>
                <a href="/family" data-route="/family">Family</a>
                <a href="/integrations" data-route="/integrations">Integrations</a>
                <div class="menu-divider"></div>
                <div id="auth-controls">
                  <div id="auth-status" class="auth-status"></div>
                  <button id="downgrade-btn" class="auth-menu-btn" style="display: none;">
                    Switch to Family Mode
                  </button>
                  <button id="upgrade-btn" class="auth-menu-btn" style="display: none;">
                    Switch to Personal Mode
                  </button>
                  <button id="logout-btn" class="auth-menu-btn logout">Sign Out</button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </nav>
    `;

    this.addStyles();
  }

  private addStyles(): void {
    const styles = `
      <style id="navigation-styles">
        .main-nav {
          background: #1f2937;
          color: white;
          padding: 0 1rem;
          box-shadow: 0 2px 4px rgba(0,0,0,0.1);
          position: sticky;
          top: 0;
          z-index: 100;
        }

        .nav-container {
          display: flex;
          justify-content: space-between;
          align-items: center;
          max-width: 1200px;
          margin: 0 auto;
          height: 3.5rem;
        }

        .nav-title {
          font-size: 1.5rem;
          font-weight: bold;
          margin: 0;
        }

        .nav-title a {
          color: white;
          text-decoration: none;
        }

        .nav-title a:hover {
          color: #e5e7eb;
        }

        .dropdown {
          position: relative;
        }

        .dropdown-btn {
          background: #374151;
          color: white;
          border: none;
          padding: 0.5rem 1rem;
          border-radius: 0.375rem;
          cursor: pointer;
          font-size: 0.875rem;
          transition: background-color 0.2s;
        }

        .dropdown-btn:hover {
          background: #4b5563;
        }

        .dropdown-content {
          position: absolute;
          right: 0;
          top: 100%;
          background: white;
          min-width: 200px;
          box-shadow: 0 10px 25px rgba(0,0,0,0.15);
          border-radius: 0.375rem;
          z-index: 1000;
          margin-top: 0.25rem;
          opacity: 0;
          visibility: hidden;
          transform: translateY(-10px);
          transition: all 0.2s ease;
        }

        .dropdown.open .dropdown-content {
          opacity: 1;
          visibility: visible;
          transform: translateY(0);
        }

        .dropdown-content a {
          display: block;
          padding: 0.75rem 1rem;
          color: #374151;
          text-decoration: none;
          border-bottom: 1px solid #f3f4f6;
          transition: background-color 0.2s;
        }

        .dropdown-content a:hover {
          background: #f9fafb;
        }

        .dropdown-content a.active {
          background: #3b82f6;
          color: white;
        }

        .dropdown-content a:last-of-type {
          border-bottom: none;
        }

        .menu-divider {
          height: 1px;
          background-color: #e5e7eb;
          margin: 0.5rem 0;
        }

        .auth-status {
          padding: 0.5rem 1rem;
          font-size: 0.75rem;
          color: #6b7280;
          border-bottom: 1px solid #f3f4f6;
          margin-bottom: 0.5rem;
        }

        .auth-status.personal {
          background-color: #dcfce7;
          color: #166534;
        }

        .auth-status.family {
          background-color: #fef3c7;
          color: #92400e;
        }

        .auth-menu-btn {
          width: 100%;
          padding: 0.5rem 1rem;
          border: none;
          background: none;
          color: #374151;
          text-align: left;
          cursor: pointer;
          font-size: 0.875rem;
          transition: background-color 0.2s;
        }

        .auth-menu-btn:hover {
          background-color: #f9fafb;
        }

        .auth-menu-btn.logout {
          color: #dc2626;
          border-top: 1px solid #f3f4f6;
          margin-top: 0.5rem;
        }

        .auth-menu-btn.logout:hover {
          background-color: #fef2f2;
        }

        /* Responsive */
        @media (max-width: 768px) {
          .nav-container {
            padding: 0 0.5rem;
          }

          .nav-title {
            font-size: 1.25rem;
          }

          .dropdown-content {
            min-width: 180px;
          }
        }
      </style>
    `;

    // Remove existing styles if any
    const existingStyles = document.getElementById('navigation-styles');
    if (existingStyles) {
      existingStyles.remove();
    }

    // Add new styles to head
    document.head.insertAdjacentHTML('beforeend', styles);
  }

  private setupEventListeners(): void {
    // Dropdown toggle
    const dropdownBtn = document.getElementById('nav-dropdown-btn');
    if (dropdownBtn) {
      dropdownBtn.addEventListener('click', () => {
        this.toggleDropdown();
      });
    }

    // Close dropdown when clicking outside
    document.addEventListener('click', e => {
      const dropdown = this.container.querySelector('.dropdown');
      if (dropdown && !dropdown.contains(e.target as Node)) {
        this.closeDropdown();
      }
    });

    // Auth buttons
    const downgradeBtn = document.getElementById('downgrade-btn');
    const upgradeBtn = document.getElementById('upgrade-btn');
    const logoutBtn = document.getElementById('logout-btn');

    if (downgradeBtn) {
      downgradeBtn.addEventListener('click', () => this.handleDowngrade());
    }

    if (upgradeBtn) {
      upgradeBtn.addEventListener('click', () => this.handleUpgrade());
    }

    if (logoutBtn) {
      logoutBtn.addEventListener('click', () => this.handleLogout());
    }

    // Close dropdown when navigation links are clicked
    const navLinks = this.container.querySelectorAll('[data-route]');
    navLinks.forEach(link => {
      link.addEventListener('click', () => {
        this.closeDropdown();
      });
    });

    // Update active state when route changes
    document.addEventListener('route-changed', () => {
      this.updateActiveRoute();
    });
  }

  private toggleDropdown(): void {
    const dropdown = this.container.querySelector('.dropdown');
    const btn = document.getElementById('nav-dropdown-btn');

    if (dropdown && btn) {
      const isOpen = dropdown.classList.contains('open');
      dropdown.classList.toggle('open', !isOpen);
      btn.setAttribute('aria-expanded', (!isOpen).toString());
      this.isDropdownOpen = !isOpen;
    }
  }

  private closeDropdown(): void {
    const dropdown = this.container.querySelector('.dropdown');
    const btn = document.getElementById('nav-dropdown-btn');

    if (dropdown && btn) {
      dropdown.classList.remove('open');
      btn.setAttribute('aria-expanded', 'false');
      this.isDropdownOpen = false;
    }
  }

  private async handleDowngrade(): Promise<void> {
    try {
      const success = await this.authManager.downgrade();
      if (success) {
        this.updateAuthState();
        // Dispatch event for other components to react
        document.dispatchEvent(new CustomEvent('auth-state-changed'));
      }
    } catch (error) {
      logger.error('Auth downgrade failed:', error);
    }
  }

  private async handleUpgrade(): Promise<void> {
    // Show password prompt
    const password = prompt('Enter your password to switch to Personal Mode:');
    if (!password) return;

    try {
      const success = await this.authManager.upgrade(password);
      if (success) {
        this.updateAuthState();
        // Dispatch event for other components to react
        document.dispatchEvent(new CustomEvent('auth-state-changed'));
      } else {
        alert('Invalid password');
      }
    } catch (error) {
      logger.error('Auth upgrade failed:', error);
      alert('Failed to switch mode');
    }
  }

  private async handleLogout(): Promise<void> {
    try {
      await this.authManager.logout();
      this.router.navigate('/login');
    } catch (error) {
      logger.error('Logout failed:', error);
    }
  }

  public updateAuthState(): void {
    const authStatus = document.getElementById('auth-status');
    const downgradeBtn = document.getElementById('downgrade-btn');
    const upgradeBtn = document.getElementById('upgrade-btn');

    if (!authStatus || !downgradeBtn || !upgradeBtn) {
      return;
    }

    const session = this.authManager.getCurrentSession();
    const user = this.authManager.getCurrentUser();

    if (session && user) {
      const isShared = session.role === 'shared';

      authStatus.textContent = isShared
        ? `Family Mode (${user.name})`
        : `Personal Mode (${session.role})`;

      authStatus.className = isShared ? 'auth-status family' : 'auth-status personal';

      downgradeBtn.style.display = isShared ? 'none' : 'block';
      upgradeBtn.style.display = isShared ? 'block' : 'none';
    } else {
      authStatus.textContent = 'Not authenticated';
      authStatus.className = 'auth-status';
      downgradeBtn.style.display = 'none';
      upgradeBtn.style.display = 'none';
    }
  }

  private updateActiveRoute(): void {
    const currentRoute = this.router.getCurrentRoute();
    if (!currentRoute) return;

    // Update navigation active states
    const navLinks = this.container.querySelectorAll('[data-route]');
    navLinks.forEach(link => {
      const linkRoute = link.getAttribute('data-route');
      if (linkRoute === currentRoute.path) {
        link.classList.add('active');
      } else {
        link.classList.remove('active');
      }
    });
  }

  destroy(): void {
    // Clean up event listeners
    const styles = document.getElementById('navigation-styles');
    if (styles) {
      styles.remove();
    }
  }
}
