# Calendar Integration Skeleton

This directory contains the skeleton structure for Google Calendar OAuth integration and sync functionality.

## Structure Created

### 1. OAuth System (`internal/oauth/`)
- `types.go` - OAuth token and configuration types
- `service.go` - OAuth flow management (auth URL generation, callback handling, token refresh)

### 2. Google Calendar Client (`internal/calendar/`)
- `google_client.go` - Google Calendar API client for fetching events and calendars

### 3. Background Sync Job (`internal/jobs/`)
- `calendar_sync.go` - Background job for periodic calendar synchronization

### 4. Database Schema (`internal/database/migrations/`)
- `014_add_oauth_integrations.sql` - OAuth tokens and sync settings tables

### 5. Web Interface
- `web/templates/calendar-settings.html.tmpl` - OAuth integration UI
- `internal/handlers/oauth_handlers.go` - OAuth flow HTTP handlers

## Next Steps for Implementation

### 1. Add Google Libraries to go.mod
```bash
go get golang.org/x/oauth2
go get google.golang.org/api/calendar/v3
go get google.golang.org/api/option
```

### 2. Update OAuth Service
Replace manual HTTP calls with `golang.org/x/oauth2` library:
- Use `oauth2.Config` for authorization flows
- Automatic token refresh handling
- Better security with PKCE support

### 3. Update Google Calendar Client
Replace manual API calls with `google.golang.org/api/calendar/v3`:
- Use official `calendar.Service`
- Proper error handling and pagination
- Type safety with generated structs

### 4. Add Encryption Integration
Update OAuth service to use existing encryption service:
- Encrypt `access_token` and `refresh_token` before database storage
- Decrypt when retrieving tokens for API calls

### 5. Wire Up Job System
- Register `CalendarSyncHandler` with job system
- Fix job system integration issues
- Add periodic scheduling for automatic sync

### 6. Add Route Registration
Add OAuth routes to server configuration:
```go
// In server setup
mux.HandleFunc("/oauth/google/connect", oauthHandlers.HandleGoogleConnect)
mux.HandleFunc("/oauth/google/callback", oauthHandlers.HandleGoogleCallback)
mux.HandleFunc("/calendar-settings", oauthHandlers.HandleCalendarSettings)
```

### 7. Environment Configuration
Add Google OAuth configuration:
```bash
GOOGLE_OAUTH_CLIENT_ID=your_client_id
GOOGLE_OAUTH_CLIENT_SECRET=your_client_secret
GOOGLE_OAUTH_REDIRECT_URL=http://localhost:8080/oauth/google/callback
```

## Benefits of Using Official Libraries

1. **OAuth2 Library**: Handles token refresh, expiration, and security automatically
2. **Google Calendar API**: Type-safe, well-documented, handles pagination and rate limiting
3. **Maintenance**: Libraries are maintained by Google and Go team
4. **Security**: Built-in protection against common OAuth vulnerabilities

## Current Status

✅ **Skeleton Structure**: Complete foundation for OAuth integration
❌ **Implementation**: Requires switching to official libraries and wiring up components
❌ **Testing**: Need to add unit tests and integration tests
❌ **Configuration**: Need environment-based configuration management

The skeleton provides a solid foundation that can be easily converted to use official Google libraries for production-ready calendar integration.