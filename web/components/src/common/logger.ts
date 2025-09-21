/**
 * Logger utility for consistent logging
 * Replaces console.log statements with proper logging levels
 */

export enum LogLevel {
  DEBUG = 0,
  INFO = 1,
  WARN = 2,
  ERROR = 3,
}

class Logger {
  private level: LogLevel = LogLevel.INFO;
  private isDevelopment = this.detectDevelopmentMode();

  setLevel(level: LogLevel): void {
    this.level = level;
  }

  private detectDevelopmentMode(): boolean {
    // Check multiple indicators for development mode
    return (
      // Check hostname for development
      (typeof window !== 'undefined' &&
        (window.location.hostname === 'localhost' ||
          window.location.hostname === '127.0.0.1' ||
          window.location.hostname.includes('dev') ||
          window.location.port !== '')) ||
      // Check for debug flag in URL
      (typeof window !== 'undefined' && window.location.search.includes('debug=true')) ||
      // Check for development indicator in global scope
      (typeof window !== 'undefined' && (window as any).DEBUG) ||
      // Default to false for production safety
      false
    );
  }

  private log(level: LogLevel, message: string, ...args: any[]): void {
    if (level < this.level) return;

    const timestamp = new Date().toISOString();
    const prefix = `[${timestamp}]`;

    switch (level) {
      case LogLevel.DEBUG:
        if (this.isDevelopment) {
          // eslint-disable-next-line no-console
          console.debug(`${prefix} DEBUG:`, message, ...args);
        }
        break;
      case LogLevel.INFO:
        if (this.isDevelopment) {
          // eslint-disable-next-line no-console
          console.info(`${prefix} INFO:`, message, ...args);
        }
        break;
      case LogLevel.WARN:
        // eslint-disable-next-line no-console
        console.warn(`${prefix} WARN:`, message, ...args);
        break;
      case LogLevel.ERROR:
        // eslint-disable-next-line no-console
        console.error(`${prefix} ERROR:`, message, ...args);
        break;
    }
  }

  debug(message: string, ...args: any[]): void {
    this.log(LogLevel.DEBUG, message, ...args);
  }

  info(message: string, ...args: any[]): void {
    this.log(LogLevel.INFO, message, ...args);
  }

  warn(message: string, ...args: any[]): void {
    this.log(LogLevel.WARN, message, ...args);
  }

  error(message: string, ...args: any[]): void {
    this.log(LogLevel.ERROR, message, ...args);
  }

  // Convenience methods for common patterns
  loadError(component: string, error: unknown): void {
    this.error(`Failed to load ${component}:`, error);
  }

  styleError(component: string, error: unknown): void {
    this.error(`Failed to load ${component} styles:`, error);
  }

  apiError(endpoint: string, error: unknown): void {
    this.error(`API request failed for ${endpoint}:`, error);
  }

  initInfo(component: string): void {
    this.info(`Initializing ${component}`);
  }

  authInfo(action: string, details?: any): void {
    this.info(`Auth: ${action}`, details);
  }

  routeInfo(route: string, details?: any): void {
    this.info(`Route: ${route}`, details);
  }

  isDevelopmentMode(): boolean {
    return this.isDevelopment;
  }
}

// Export singleton instance
export const logger = new Logger();

// Set development level if in development mode
if (logger.isDevelopmentMode()) {
  logger.setLevel(LogLevel.DEBUG);
}
