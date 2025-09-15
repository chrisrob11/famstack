# Fam-Stack

[![CI](https://github.com/famstack/famstack/actions/workflows/ci.yml/badge.svg)](https://github.com/famstack/famstack/actions/workflows/ci.yml)
[![Release](https://github.com/famstack/famstack/actions/workflows/release.yml/badge.svg)](https://github.com/famstack/famstack/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/famstack/famstack)](https://goreportcard.com/report/github.com/famstack/famstack)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

An open-source family task management system that helps families organize todos, chores, and appointments in one unified platform. Built for self-hosting with a focus on simplicity and ease of deployment.

## Features

- **Unified Task System**: Manage todos, chores, and appointments in one place
- **Family Member Management**: Role-based access for parents and children
- **Interactive Web UI**: Server-rendered HTML with HTMX for dynamic interactions
- **TypeScript Components**: Rich interactions like drag-and-drop task reordering
- **Single Binary Deployment**: No external dependencies required
- **SQLite Database**: Embedded database with automatic migrations
- **Session-based Authentication**: Simple, secure authentication system

## Quick Start

### Download and Run

1. Download the latest release for your platform from [GitHub Releases](https://github.com/famstack/famstack/releases)
2. Make it executable: `chmod +x famstack-*`
3. Run it: `./famstack-*`
4. Open your browser to `http://localhost:8080`

### Build from Source

```bash
# Clone the repository
git clone https://github.com/famstack/famstack.git
cd famstack

# Install build tools
make install-tools

# Build the application
make build

# Run the application
make run
```

## Development

### Prerequisites

- Go 1.24.7
- Node.js 23.2.0
- Make

### Setup

#### Option 1: Using asdf (Recommended)

If you have [asdf](https://asdf-vm.com/) installed, the project includes a `.tool-versions` file for consistent tool versions:

```bash
# Clone the repository
git clone https://github.com/famstack/famstack.git
cd famstack

# Install the exact tool versions used by the project
asdf install

# Install all required tools
make install-tools

# Run in development mode
make dev
```

#### Option 2: Manual Installation

```bash
# Clone the repository
git clone https://github.com/famstack/famstack.git
cd famstack

# Ensure you have the right versions:
# - Go 1.24.7
# - Node.js 23.2.0

# Install all required tools
make install-tools

# Run in development mode
make dev
```

### Available Make Targets

- `make build` - Build TypeScript components and Go binary
- `make run` - Run the application locally
- `make test` - Run all tests (Go + TypeScript)
- `make lint` - Run all linters (golangci-lint + ESLint)
- `make clean` - Clean build artifacts
- `make dev` - Development mode with file watching
- `make help` - Show all available targets

### Project Structure

```
fam-stack/
├── cmd/famstack/           # Main executable
├── internal/               # Private application code
│   ├── auth/              # Authentication logic
│   ├── database/          # Database setup, migrations
│   ├── handlers/          # HTTP handlers
│   ├── models/            # Data models
│   └── server/            # Server setup
├── web/                   # Web assets (embedded)
│   ├── templates/         # HTML templates
│   ├── static/            # CSS and compiled JS
│   └── components/        # TypeScript source files
├── migrations/            # SQL migration files
└── .github/workflows/     # CI/CD workflows
```

## Configuration

### Command Line Options

```bash
./famstack [options]

Options:
  -port string
        Server port (default "8080")
  -db string
        Database file path (default "famstack.db")
  -migrate-up
        Run database migrations up
  -migrate-down
        Run database migrations down
  -dev
        Enable development mode
```

### Environment Variables

- `PORT` - Server port (overrides -port flag)
- `DATABASE_PATH` - Database file path (overrides -db flag)

## Database

Fam-Stack uses SQLite with the following schema:

- **families** - Family units
- **users** - Family members with roles (parent/child/admin)
- **tasks** - Unified task system (todos/chores/appointments)
- **sessions** - User authentication sessions

### Migrations

Database migrations are handled automatically on startup, but you can also run them manually:

```bash
# Run migrations up
./famstack -migrate-up

# Rollback one migration
./famstack -migrate-down
```

## API

Fam-Stack provides both a web interface and REST API endpoints:

### ✅ Currently Working API Endpoints

**Tasks (Fully Implemented)**
- `GET /api/tasks` - List tasks with user grouping ✅
- `POST /api/tasks` - Create new task ✅
- `PATCH /api/tasks/:id` - Update task (status, assignment, title) ✅
- `DELETE /api/tasks/:id` - Delete task ✅
- `GET /api/tasks/:id` - Get single task ✅

**Family Management (Fully Implemented)**
- `GET /api/families` - List all families ✅
- `POST /api/families` - Create new family ✅
- `GET /api/families/:id` - Get family details ✅
- `GET /api/family/members` - List family members ✅
- `POST /api/family/members` - Add family member ✅
- `PATCH /api/family/members/:id` - Update family member ✅
- `DELETE /api/family/members/:id` - Delete family member ✅

**Calendar Infrastructure (Database & API Complete)**
- `GET /api/calendar` - Get unified calendar events ✅
- `POST /api/calendar` - Create new calendar event ✅
- `PATCH /api/calendar/:id` - Update calendar event ✅
- `DELETE /api/calendar/:id` - Delete calendar event ✅

**Schedules (Fully Implemented)**
- `GET /api/schedules` - List recurring schedules ✅
- `POST /api/schedules` - Create new schedule ✅

### ❌ Missing Authentication Endpoints
- `POST /auth/login` - User login ❌
- `POST /auth/logout` - User logout ❌
- `POST /auth/register` - User registration ❌

### ❌ Missing Calendar Integration Endpoints
- `POST /auth/oauth/google` - Google Calendar OAuth flow ❌
- `POST /auth/oauth/microsoft` - Microsoft Outlook OAuth flow ❌
- `POST /api/calendar/sync` - Trigger calendar data sync ❌
- `GET /api/calendar/providers` - List connected calendar providers ❌

**Note**: Core API functionality and calendar infrastructure work perfectly, but currently operate without authentication. External calendar integration (OAuth + sync) is not implemented - the system has sample calendar events and full CRUD operations, but can't import from Google/Outlook/Apple calendars.

## TypeScript Components

The application includes TypeScript components for enhanced interactivity:

- **TaskManager**: Drag-and-drop task reordering, inline editing
- **FamilySelector**: Multi-select family member assignment
- **Real-time Updates**: HTMX integration for dynamic content

Components are automatically initialized and integrate seamlessly with server-rendered HTML.

## Deployment

### Single Binary

Fam-Stack compiles to a single binary with all assets embedded. Simply copy the binary to your server and run it:

```bash
# Download and run
wget https://github.com/famstack/famstack/releases/latest/download/famstack-linux-amd64.tar.gz
tar -xzf famstack-linux-amd64.tar.gz
./famstack
```

### Docker

```bash
# Build Docker image
docker build -t famstack .

# Run container
docker run -p 8080:8080 -v $(pwd)/data:/data famstack -db /data/famstack.db
```

### Systemd Service

Create `/etc/systemd/system/famstack.service`:

```ini
[Unit]
Description=Fam-Stack Family Task Manager
After=network.target

[Service]
Type=simple
User=famstack
WorkingDirectory=/opt/famstack
ExecStart=/opt/famstack/famstack -port 8080 -db /var/lib/famstack/famstack.db
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes following our coding standards
4. Write conventional commit messages (see below)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

### Commit Message Convention

This project uses [Conventional Commits](https://www.conventionalcommits.org/) for automatic semantic versioning and changelog generation.

**Format:** `<type>(<scope>): <description>`

**Types:**
- `feat:` - A new feature (minor version bump)
- `fix:` - A bug fix (patch version bump)  
- `perf:` - Performance improvements (patch version bump)
- `revert:` - Revert a previous commit (patch version bump)
- `docs:` - Documentation only changes (no version bump)
- `style:` - Code style changes (no version bump)
- `refactor:` - Code refactoring (no version bump)
- `test:` - Adding or updating tests (no version bump)
- `chore:` - Build process or tooling changes (no version bump)
- `ci:` - CI/CD changes (no version bump)
- `build:` - Build system changes (no version bump)

**Examples:**
```bash
feat(auth): add session-based authentication
fix(database): resolve migration rollback issue
docs(readme): update installation instructions
chore(deps): update go dependencies
```

**Breaking Changes:** Add `BREAKING CHANGE:` in the footer for major version bumps.

### Development Workflow

- Run `make lint` before committing
- Ensure all tests pass with `make test`  
- Follow conventional commit message format
- Use the provided `.gitmessage` template: `git config commit.template .gitmessage`
- Follow Go and TypeScript best practices
- Update documentation as needed

### Automatic Releases

- Commits to `main` branch trigger automatic semantic versioning
- Releases are created automatically with cross-platform binaries
- CHANGELOG.md is updated automatically
- Version tags follow semver (v1.2.3)

## Development Status

### 🚀 Current Working Features

**End-to-End Task Management:**
- ✅ Task list display with real sample data
- ✅ Color-coded task types (todos=blue, chores=orange, appointments=purple)
- ✅ Task status indicators (pending/completed)
- ✅ Family member assignments visible
- ✅ Responsive web UI with clean styling
- ✅ SQLite database with automatic migrations
- ✅ Sample data (Smith family with 4 tasks)

**Getting Started:**
```bash
make dev-setup    # Reset and setup fresh environment
make run          # Start server on http://localhost:8080
```

Visit `http://localhost:8080` to see the working task dashboard!

### Current Implementation State

**✅ Fully Implemented & Working:**
- **✅ Database Layer**: Complete SQLite database with 10 migrations applied, all schemas working
- **✅ Data Models**: Full models for families, users, tasks, sessions, jobs, and schedules
- **✅ Task Management API**: Complete CRUD operations for tasks with full REST endpoints
- **✅ Family Management API**: Complete user and family management with role-based access
- **✅ Calendar Infrastructure**: Complete calendar database schema and API layer
- **✅ Frontend Components**: Rich TypeScript components with drag-and-drop, modals, and interactions
- **✅ Background Job System**: Enterprise-grade job system with optimistic concurrency control
- **✅ Task Scheduling**: Automated task generation and recurring schedule management
- **✅ HTML Templates**: Responsive multi-page application with navigation
- **✅ HTMX Integration**: Dynamic page updates and API interactions
- **✅ Build System**: Complete TypeScript compilation and asset embedding
- **✅ Database Migrations**: Automatic migration system with rollback support
- **✅ Sample Data**: Working Smith family with real task data for demonstration

**❌ Missing Critical Components:**

*Authentication & Security:*
- **❌ Authentication System**: No login/logout/registration handlers (internal/auth/ is empty)
- **❌ Session Management**: No session middleware or authentication checks
- **❌ Security Middleware**: No CSRF protection, input validation, or authorization
- **❌ Password Management**: No password hashing, validation, or reset functionality
- **❌ Access Control**: No role-based authorization enforcement in API endpoints

*Calendar Integration:*
- **❌ OAuth Flow**: No Google/Microsoft/Apple Calendar OAuth authentication
- **❌ External Calendar Sync**: No calendar data import from external providers
- **❌ Calendar API Clients**: No Google Calendar API, Outlook API, or CalDAV integration
- **❌ Data Pipeline**: No sync jobs to transform external calendar data into unified events
- **❌ Sync Management**: No conflict resolution, deletion handling, or sync scheduling

### Next Development Priorities

**Immediate Priority - Authentication & Security:**

1. **Authentication System** (`internal/auth/`)
   - Implement session-based authentication middleware
   - Create login/logout/registration HTTP handlers
   - Add password hashing (bcrypt) and validation
   - Build user registration and login forms

2. **Security Middleware** (`internal/middleware/`)
   - CSRF protection for all forms and API endpoints
   - Role-based authorization middleware (parent/child/admin access)
   - Input validation and sanitization middleware
   - Secure session cookie configuration

3. **Access Control Integration**
   - Protect all API endpoints with authentication checks
   - Enforce family-based data isolation (users can only see their family's data)
   - Add role-based permissions for task creation/editing

**Secondary Priority - Calendar Integration:**

4. **OAuth Integration** (`internal/oauth/`)
   - Google Calendar OAuth 2.0 flow implementation
   - Microsoft Graph API OAuth for Outlook integration
   - Apple Calendar (CalDAV) authentication
   - Token storage, refresh, and management

5. **Calendar Data Pipeline** (`internal/calendar/`)
   - Google Calendar API client for event fetching
   - Microsoft Graph API client for Outlook events
   - CalDAV client for Apple/other calendar systems
   - Background jobs for periodic calendar sync

6. **Data Transformation** (`internal/jobs/`)
   - Import raw calendar events into `calendar_events` table
   - Transform and merge into `unified_calendar_events` for UI
   - Handle recurring events, time zones, and event updates
   - Conflict resolution for overlapping events

**Future Enhancements:**
7. **Advanced Features**
   - Password reset functionality
   - Email verification system
   - Multi-factor authentication
   - API rate limiting and throttling
   - Real-time calendar sync webhooks
   - Smart event categorization and tagging

**Note**: The core application functionality (task management, family management, frontend components, background jobs, calendar infrastructure) is fully implemented and working. The missing pieces are authentication/security and external calendar integration - once these are added, the application will be production-ready.

### Technology Stack

**Backend:**
- Go 1.23 with standard library
- SQLite with CGO for database
- Goose for database migrations
- Embedded assets for single binary deployment

**Frontend:**
- TypeScript with modern ES6+ features
- SortableJS for drag-and-drop functionality
- HTMX for dynamic server interactions
- CSS for responsive styling

**Development Tools:**
- Make for build automation
- golangci-lint for Go code quality
- ESLint for TypeScript linting
- GitHub Actions for CI/CD

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with Go standard library and minimal dependencies
- Uses HTMX for dynamic web interactions
- TypeScript for rich client-side components
- SQLite for embedded database storage