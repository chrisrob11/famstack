import { logger } from './logger.js';

export interface ErrorContext {
  component: string;
  operation: string;
  details?: any;
}

export class ComponentError extends Error {
  constructor(
    message: string,
    public context: ErrorContext,
    public originalError?: Error
  ) {
    super(message);
    this.name = 'ComponentError';
  }
}

export const errorHandler = {
  /**
   * Handle async operations with consistent error logging and reporting
   */
  async handleAsync<T>(
    operation: () => Promise<T>,
    context: ErrorContext,
    fallbackValue?: T
  ): Promise<T | undefined> {
    try {
      return await operation();
    } catch (error) {
      logger.error(`Error in ${context.component}.${context.operation}:`, error);

      const componentError = new ComponentError(
        `Failed to ${context.operation}`,
        context,
        error instanceof Error ? error : new Error(String(error))
      );

      if (fallbackValue !== undefined) {
        return fallbackValue;
      }

      throw componentError;
    }
  },

  /**
   * Handle synchronous operations with error logging
   */
  handleSync<T>(
    operation: () => T,
    context: ErrorContext,
    fallbackValue?: T
  ): T | undefined {
    try {
      return operation();
    } catch (error) {
      logger.error(`Error in ${context.component}.${context.operation}:`, error);

      if (fallbackValue !== undefined) {
        return fallbackValue;
      }

      throw new ComponentError(
        `Failed to ${context.operation}`,
        context,
        error instanceof Error ? error : new Error(String(error))
      );
    }
  },

  /**
   * Log and dispatch error events for components
   */
  dispatchError(
    element: HTMLElement,
    eventName: string,
    error: Error | ComponentError,
    detail?: any
  ): void {
    const errorDetail = {
      error: error.message,
      stack: error.stack,
      ...detail
    };

    logger.error(`Dispatching error event ${eventName}:`, errorDetail);

    element.dispatchEvent(
      new CustomEvent(eventName, {
        detail: errorDetail,
        bubbles: true,
      })
    );
  }
};