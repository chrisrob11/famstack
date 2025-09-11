// Jest setup file
// Mock global objects that aren't available in jsdom
global.fetch = jest.fn();

// Mock console to reduce noise during testing
global.console = {
  ...console,
  log: jest.fn(),
  debug: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
  error: jest.fn(),
};