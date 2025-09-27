/**
 * Shared time formatting utilities for calendar components
 */

export interface TimeSlot {
  hour: number;
  minutes: number;
  timeString: string;
  isHourBoundary: boolean;
  position: number;
  totalMinutes: number;
}

export class TimeFormatter {
  constructor(private use24Hour: boolean = false) {}

  /**
   * Format a time slot for display in the calendar grid
   */
  formatTimeSlot(hour: number, minutes: number): string {
    if (this.use24Hour) {
      // 24-hour format
      const hourStr = hour.toString().padStart(2, '0');
      return minutes === 0
        ? `${hourStr}:00`
        : `${hourStr}:${minutes.toString().padStart(2, '0')}`;
    } else {
      // 12-hour format
      const period = hour >= 12 ? 'PM' : 'AM';
      const displayHour = hour > 12 ? hour - 12 : hour === 0 ? 12 : hour;
      return minutes === 0
        ? `${displayHour} ${period}`
        : `${displayHour}:${minutes.toString().padStart(2, '0')} ${period}`;
    }
  }

  /**
   * Format a time range for display in event cards
   */
  formatTimeRange(startTime: Date, endTime: Date): string {
    const formatTime = (date: Date) => {
      if (this.use24Hour) {
        return date.toLocaleTimeString('en-US', {
          hour: '2-digit',
          minute: '2-digit',
          hour12: false
        });
      } else {
        return date.toLocaleTimeString('en-US', {
          hour: 'numeric',
          minute: '2-digit',
          hour12: true
        });
      }
    };

    return `${formatTime(startTime)} - ${formatTime(endTime)}`;
  }

  /**
   * Format a date for display in the calendar header
   */
  formatDate(dateString: string | undefined): string {
    if (!dateString) {
      dateString = new Date().toISOString().split('T')[0];
    }
    const date = new Date(dateString!);
    return date.toLocaleDateString(undefined, {
      weekday: 'long',
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  }

  /**
   * Generate time slots for the calendar grid
   */
  generateTimeSlots(pixelPerMinute: number, startHour = 0, endHour = 24): TimeSlot[] {
    const slots: TimeSlot[] = [];

    for (let hour = startHour; hour < endHour; hour++) {
      for (let quarter = 0; quarter < 4; quarter++) {
        const minutes = quarter * 15;
        const isHourBoundary = quarter === 0;
        const timeString = this.formatTimeSlot(hour, minutes);
        const totalMinutes = hour * 60 + minutes;
        const position = totalMinutes * pixelPerMinute;

        slots.push({
          hour,
          minutes,
          timeString,
          isHourBoundary,
          position,
          totalMinutes,
        });
      }
    }

    return slots;
  }
}