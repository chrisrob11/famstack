# Authentication System Implementation

## Overview

This document describes the comprehensive JWT-based authentication system implemented for FamStack, featuring secure role-based access control with permission downgrade/upgrade capabilities.

## Architecture

### Core Components

1. **JWT Token Management** (`internal/auth/jwt.go`)
   - Stateless JWT tokens with HS256 signing
   - Token creation, validation, refresh, and expiration checking
   - Support for role downgrade/upgrade while maintaining original permissions

2. **User Management** (`internal/auth/types.go`)
   - Enhanced user model with `first_name`, `last_name`, `nickname`
   - Email verification and last login tracking
   - Support for multiple display name preferences

3. **Permission System** (`internal/auth/permissions.go`)
   - Entity-Action-Scope permission model
   - Three role levels: `shared`, `user`, `admin`
   - Field-level governance for task updates

4. **Password Security** (`internal/auth/password.go`)
   - Argon2id password hashing (industry standard)
   - Configurable memory, iterations, and parallelism
   - Secure password validation and requirements

5. **HTTP Middleware** (`internal/auth/middleware.go`)
   - Authentication middleware for protected routes
   - Permission-based route protection
   - Optional authentication for public routes

6. **API Handlers** (`internal/auth/handlers.go`)
   - Login/logout endpoints
   - Role downgrade/upgrade endpoints
   - Token refresh and user info endpoints

## Permission Model

### Roles

- **Shared Mode** (`shared`): Minimal permissions for family kiosk use
  - Read tasks and calendar
  - Update task status only (mark complete/incomplete)
  - No create, delete, or administrative actions

- **User Mode** (`user`): Standard family member permissions
  - Full task management (create, read, update own, delete own)
  - Calendar management (create events, update/delete own)
  - View family information
  - No administrative functions

- **Admin Mode** (`admin`): Full administrative permissions
  - All user permissions plus:
  - Delete any tasks/events
  - Manage family members
  - Access settings and analytics
  - Full system administration

### Permission Format

Permissions follow the pattern: `entity:action:scope`

- **Entities**: `task`, `calendar`, `family`, `user`, `setting`, `analytic`
- **Actions**: `create`, `read`, `update`, `delete`
- **Scopes**: `none`, `own`, `any`

Examples:
- `task:read:any` - Can read all tasks
- `task:update:own` - Can update only own tasks
- `task:delete:none` - Cannot delete any tasks

## Authentication Flow

### Initial Login
```
1. User submits email/password
2. Server validates credentials with Argon2id
3. JWT token created with user's role (7-day expiration)
4. Token stored in HTTP-only cookie
5. User has full permissions for their role
```

### Downgrade to Shared Mode
```
1. Authenticated user clicks "Switch to Family Mode"
2. New JWT created with role=shared, original_role=user
3. Limited permissions activated
4. Same token expiration maintained
```

### Upgrade from Shared Mode
```
1. User in shared mode attempts restricted action
2. Password challenge modal appears
3. User enters password
4. Server validates password with rate limiting (5 attempts/15min)
5. New JWT created with role=original_role
6. Full permissions restored
```

## Database Schema Updates

### Migration 011: Authentication Enhancement

```sql
-- Add user name fields
ALTER TABLE users ADD COLUMN first_name TEXT;
ALTER TABLE users ADD COLUMN last_name TEXT;
ALTER TABLE users ADD COLUMN nickname TEXT;
ALTER TABLE users ADD COLUMN email_verified BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN last_login_at DATETIME;

-- Sample users with secure password hashes
-- Password: "password123" for all sample users
UPDATE users SET
    password_hash = '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LeYQjNjJIGKyq2/OK',
    first_name = 'John|Jane|Bobby',
    last_name = 'Smith',
    email_verified = TRUE;
```

## API Endpoints

### Authentication Routes

- `POST /auth/login` - User login with email/password
- `POST /auth/logout` - Clear authentication cookie
- `POST /auth/downgrade` - Switch to family shared mode
- `POST /auth/upgrade` - Upgrade with password challenge
- `POST /auth/refresh` - Refresh token expiration
- `GET /auth/me` - Get current user info and permissions

### Route Protection

```go
// Require authentication
router.Use(authMiddleware.RequireAuth)

// Require specific permissions
router.Handle("/api/tasks",
    authMiddleware.RequireEntityAction(EntityTask, ActionCreate)(handler))
```

## Security Features

1. **Secure Password Hashing**
   - Argon2id with 64MB memory, 3 iterations, parallelism=2
   - Salt length: 16 bytes, Key length: 32 bytes
   - Resistant to rainbow table and brute force attacks

2. **JWT Security**
   - HS256 signing with 256-bit secret keys
   - HTTP-only cookies prevent XSS attacks
   - 7-day expiration with refresh capability
   - Stateless design for scalability

3. **Rate Limiting**
   - Password upgrade attempts: 5 per 15 minutes per user
   - In-memory tracking with automatic cleanup

4. **Field-Level Security**
   - Shared mode can only update task status/completion
   - Users can update own content details
   - Admins can modify any field including assignments

## Development Usage

### Sample Login Credentials

All sample users have password: `password123`

- `john@smith.com` - John Smith (user role)
- `jane@smith.com` - Jane Smith (user role)
- `bobby@smith.com` - Bobby Smith (user role, nickname: "Bob")

### Testing Authentication

```bash
# Login
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@smith.com","password":"password123"}'

# Switch to family mode
curl -X POST http://localhost:8080/auth/downgrade \
  -H "Cookie: auth_token=<jwt_token>"

# Upgrade back to personal mode
curl -X POST http://localhost:8080/auth/upgrade \
  -H "Content-Type: application/json" \
  -H "Cookie: auth_token=<shared_jwt_token>" \
  -d '{"password":"password123"}'
```

## Next Steps

The authentication system backend is now complete. Future frontend implementation should:

1. Create login/logout UI components
2. Implement mode switching interface
3. Add permission-based UI element visibility
4. Create password challenge modals
5. Handle authentication state management
6. Implement automatic token refresh

This provides a solid foundation for secure, family-friendly authentication with flexible permission management.