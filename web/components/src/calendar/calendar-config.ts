/**
 * Calendar configuration constants and settings
 */

export const CALENDAR_CONFIG = {
  // Time grid configuration
  TIME_SLOT_HEIGHT: 20, // pixels per 15-minute slot
  PIXEL_PER_MINUTE: 20 / 15, // 1.33 pixels per minute

  // Layout defaults
  DEFAULT_CONTAINER_WIDTH: 400, // fallback width for layout calculations
  DEFAULT_SCROLL_HOUR: 7, // scroll to 7 AM on initial load

  // Time range
  CALENDAR_START_HOUR: 0, // midnight
  CALENDAR_END_HOUR: 24, // midnight next day
  MINUTES_PER_DAY: 24 * 60, // 1440 minutes

  // Performance
  RESIZE_DEBOUNCE_MS: 16, // ~60fps for resize events
  LAYOUT_RETRY_DELAY_MS: 100, // delay for DOM rendering

  // Accessibility
  FOCUS_OUTLINE_WIDTH: 2, // pixels
  MIN_TOUCH_TARGET: 44, // pixels for mobile accessibility
} as const;

export const CALENDAR_STYLES = {
  // Color scheme
  COLORS: {
    BACKGROUND: '#f8f9fa',
    TEXT: '#333',
    GRID_BORDER: '#e1e5e9',
    HOUR_TEXT: '#6c757d',
    HOUR_LINE: '#dee2e6',
    ERROR_BACKGROUND: '#fff3cd',
    ERROR_BORDER: '#ffeaa7',
    ERROR_TEXT: '#856404',
  },

  // Spacing
  SPACING: {
    CONTAINER_PADDING: 16, // outer calendar padding
    HEADER_PADDING: 20, // header top/bottom padding
    EVENT_GAP: 10, // gap between consecutive events
    CONTAINER_PADDING_EVENTS: 30, // padding for event containers (for rounded corners)
  },

  // Typography
  TYPOGRAPHY: {
    DATE_TITLE_SIZE: 24, // header date font size
    DATE_TITLE_SIZE_MOBILE: 20, // mobile header date font size
    TIME_LABEL_SIZE: 12, // time label font size
    EVENT_FONT_SIZE_SMALL: 11, // event text for small events
    EVENT_FONT_SIZE_LARGE: 12, // event text for large events
  },

  // Dimensions
  DIMENSIONS: {
    TIME_LABEL_WIDTH: 40, // width of time labels column
    SCROLLBAR_WIDTH: 8, // custom scrollbar width
    BORDER_RADIUS: 8, // standard border radius
    EVENT_MIN_HEIGHT: 20, // minimum event height for readability
  },
} as const;

// Responsive breakpoints
export const BREAKPOINTS = {
  MOBILE: 768,
  TABLET: 1024,
  DESKTOP: 1200,
} as const;

// Event card configuration thresholds
export const EVENT_CARD_CONFIG = {
  SHOW_CIRCLES_MIN_HEIGHT: 20, // minimum height to show person circles
  SHOW_TIME_MIN_HEIGHT: 32, // minimum height to show time info
  FONT_SIZE_THRESHOLD: 30, // height threshold for larger font
  TALL_EVENT_THRESHOLD: 50, // height threshold for column layout
} as const;
