# Calendar Integration

Code for connecting to external calendars (Google, Outlook, etc.).

## Current status

❌ **Not working yet** - The code structure exists but isn't connected

## What exists

- Basic database tables for OAuth tokens
- Skeleton code for Google Calendar API
- Background job framework for syncing

## What's missing

- Actual Google API integration
- OAuth flow implementation
- Sync logic to import events

## To implement

1. Add Google OAuth libraries to go.mod
2. Wire up the OAuth flow handlers
3. Implement event syncing from Google Calendar
4. Add error handling and rate limiting

## Files

- `google_client.go` - Google Calendar API (skeleton)
- `../oauth/` - OAuth token management (skeleton)
- `../jobs/calendar_sync.go` - Background sync job (skeleton)

The calendar system works with manual events, but external calendar sync is not implemented yet.