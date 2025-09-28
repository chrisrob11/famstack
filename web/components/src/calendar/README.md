# Calendar Component

Daily calendar view for families. Shows today's events for everyone.

## What's working

✅ **Basic calendar** - 24-hour time grid
✅ **Event display** - Shows events in correct time slots
✅ **Person identification** - Colored circles show who's attending
✅ **Timezone handling** - Proper UTC storage with local display
✅ **Date navigation** - Previous/next day, today button

## What's working now

✅ **Event overlap handling** - Fixed! Events now share horizontal space properly

## What's missing

❌ Person filtering (click person to see only their events)
❌ Tablet optimization
❌ Touch-friendly interface

## File structure

```
calendar/
├── daily-calendar.ts       # Main component
├── calendar-api.ts         # API calls
├── calendar-config.ts      # Settings
└── utils/
    ├── time-formatter.ts   # Time display
    └── layout-utils.ts     # Event positioning
```

## Development

Visit `/calendar-dev` to see the test page.

## API data format

Events need these fields:
- `id`, `title`, `start_time`, `end_time`
- `attendees` with `name`, `color`, `initial`
- Times in UTC, displayed in family timezone

## Current status

The calendar shows events correctly but they overlap visually when scheduled at the same time. Event positioning logic needs work.