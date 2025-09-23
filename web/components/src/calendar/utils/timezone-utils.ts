/**
 * Timezone utilities for calendar components
 * Handles conversion between UTC storage and local display times
 */

/**
 * Parses a UTC ISO string and returns a Date object in the local timezone
 * @param utcIsoString - ISO 8601 string in UTC (e.g., "2023-09-22T14:30:00Z")
 * @returns Date object adjusted to local timezone
 */
export function parseUTCToLocal(utcIsoString: string): Date {
  // Ensure the string has a 'Z' suffix for proper UTC parsing
  const normalizedString = utcIsoString.endsWith('Z') ? utcIsoString : utcIsoString + 'Z';

  // Parse as UTC and create a new Date in local timezone
  const utcDate = new Date(normalizedString);

  // Return the UTC date which will be displayed in local time
  return utcDate;
}

/**
 * Converts a local Date to UTC ISO string for API submission
 * @param localDate - Date object in local timezone
 * @returns ISO string in UTC format
 */
export function formatLocalToUTC(localDate: Date): string {
  return localDate.toISOString();
}

/**
 * Gets the user's timezone (browser detected)
 * @returns IANA timezone identifier (e.g., "America/New_York")
 */
export function getUserTimezone(): string {
  return Intl.DateTimeFormat().resolvedOptions().timeZone;
}

/**
 * Formats a time for display in the user's timezone
 * @param date - Date object to format
 * @param options - Intl.DateTimeFormatOptions
 * @returns Formatted time string
 */
export function formatTimeInUserTimezone(
  date: Date,
  options: Intl.DateTimeFormatOptions = {}
): string {
  const defaultOptions: Intl.DateTimeFormatOptions = {
    hour: 'numeric',
    minute: '2-digit',
    ...options
  };

  return date.toLocaleTimeString(undefined, defaultOptions);
}

/**
 * Checks if the browser supports the given timezone
 * @param timezone - IANA timezone identifier
 * @returns true if timezone is supported
 */
export function isTimezoneSupported(timezone: string): boolean {
  try {
    Intl.DateTimeFormat(undefined, { timeZone: timezone });
    return true;
  } catch {
    return false;
  }
}

/**
 * For development/debugging: logs timezone conversion details
 * @param utcString - Original UTC string
 * @param localDate - Converted local date
 */
export function debugTimezoneConversion(utcString: string, localDate: Date): void {
  if (process.env['NODE_ENV'] === 'development') {
    console.log('🕒 Timezone conversion:', {
      original: utcString,
      parsed: localDate.toISOString(),
      display: localDate.toLocaleString(),
      timezone: getUserTimezone()
    });
  }
}