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

### Authentication
- `POST /auth/login` - User login
- `POST /auth/logout` - User logout
- `POST /auth/register` - User registration

### Tasks
- `GET /api/tasks` - List tasks
- `POST /api/tasks` - Create task
- `PATCH /api/tasks/:id` - Update task
- `DELETE /api/tasks/:id` - Delete task
- `PATCH /api/tasks/:id/complete` - Mark task as complete

### Family Management
- `GET /api/family/members` - List family members
- `POST /api/family/members` - Add family member
- `PATCH /api/family/members/:id` - Update family member

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

**✅ Completed Components:**
- Basic project structure and data models
- Database layer with SQLite and migration system
- TypeScript components with rich interactions (drag-and-drop, inline editing)
- Build system with Make and embedded assets
- CLI interface with graceful shutdown
- Cross-platform build and release automation
- **✅ Database migrations** - Complete schema with sample data for families, users, tasks
- **✅ Basic HTTP handlers** - Task list display with embedded HTML templates
- **✅ HTML templates** - Responsive task dashboard with CSS styling
- **✅ Static assets** - Integrated CSS styling and TypeScript build pipeline
- **✅ End-to-end functionality** - Working task list from database to browser

**⚠️ In Progress / Missing Components:**
- [ ] Authentication system - Session management, login/logout/registration endpoints
- [ ] Authorization middleware - Role-based access control (parent/child/admin)
- [ ] CRUD operations - Task creation, editing, deletion endpoints
- [ ] API endpoints - RESTful API for task management operations
- [ ] HTMX integration - Dynamic updates without full page reloads
- [ ] TypeScript integration - Connect drag-and-drop components to backend APIs
- [ ] CSRF protection and security middleware
- [ ] Logging and error handling middleware

### Next Development Priorities

1. **Task CRUD Operations** (`internal/handlers/`)
   - Task creation form and POST handler
   - Task editing (inline and form-based)
   - Task deletion with confirmation
   - Task completion toggle

2. **Authentication System** (`internal/auth/`, `internal/handlers/`)
   - Session management middleware
   - Login/logout/registration handlers
   - Password hashing and validation
   - Role-based access control

3. **Interactive Frontend** (`web/components/`, `web/templates/`)
   - Connect existing TypeScript drag-and-drop to backend APIs
   - HTMX integration for dynamic updates
   - Real-time task reordering and editing
   - Form validation and error handling

4. **API Endpoints** (`internal/handlers/`)
   - RESTful API for task management
   - Family and user management endpoints
   - Bulk operations (assign multiple tasks)

5. **Security & Polish**
   - CSRF protection middleware
   - Input validation and sanitization
   - Logging and error handling
   - Responsive design improvements

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