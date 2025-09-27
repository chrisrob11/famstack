# Authentication

How login and user permissions work in FamStack.

## User types

**Admin** - Can do everything
**User** - Can manage their own tasks and calendar
**Shared** - Can only mark tasks as done (good for family tablets)

## How login works

1. Enter email and password
2. Get logged in for 4 hours
3. Can switch to "Family Mode" for shared device use
4. Enter password again to get back to personal mode

## Setting up users

```bash
# Create a new user
./famstack user create

# List users
./famstack user list

# Reset password
./famstack user reset-password --email user@example.com
```

## Test users

For development, these users exist with password `password123`:
- john@smith.com
- jane@smith.com
- bobby@smith.com

## API endpoints

- `POST /auth/login` - Log in
- `POST /auth/logout` - Log out
- `POST /auth/downgrade` - Switch to family mode
- `POST /auth/upgrade` - Switch back to personal mode
- `GET /auth/me` - Get current user info

## Security notes

- Passwords are hashed with Argon2id
- Sessions expire after 4 hours
- Rate limiting on password attempts
- Cookies are HTTP-only to prevent XSS

## What's implemented

- User login/logout ✅
- Role switching ✅
- Password hashing ✅
- Session management ✅

## What's missing

- Web UI for login