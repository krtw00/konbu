<p align="center">
  <img src="web/static/favicon.svg" width="64" height="64" alt="konbu">
</p>

<h1 align="center">konbu</h1>

<p align="center">Manage memos, todos, calendar, and bookmarks<br>in one place — with AI and cross-search.</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="MIT License"></a>
  <a href="https://github.com/krtw00/konbu/actions"><img src="https://github.com/krtw00/konbu/actions/workflows/deploy.yml/badge.svg" alt="Deploy"></a>
</p>

<p align="center"><a href="README.ja.md">日本語</a> | English</p>

<p align="center"><img src="docs/screenshot.png" width="800" alt="konbu screenshot"></p>

---

## Try It

- **Cloud** -- Use instantly at [konbu-cloud.codenica.dev](https://konbu-cloud.codenica.dev) (free, no setup)
- **Self-hosted** -- Run on your own server with Docker (see below)

## Why konbu?

Memos, tasks, calendar, and bookmarks are usually scattered across separate apps. Information gets fragmented, and "where did I write that?" becomes a daily frustration. konbu brings everything into one place with cross-search and AI chat.

## Features

- **Cross-search** -- Full-text search across memos, todos, events, and bookmarks. No more "where was that?"
- **AI Chat** -- "Add groceries to my todo" "What's my schedule tomorrow?" — manage everything with natural language (free tier included)
- **Memos** -- Markdown notes with tagging, live preview
- **ToDo** -- Inline task creation with due dates, tags, and notes
- **Calendar** -- Monthly view with event CRUD and iCal import
- **Bookmarks** -- Site management with categories and drag-and-drop reordering
- **CLI & MCP** -- Full-featured CLI client and MCP server for AI agent integration
- **Export/Import** -- JSON export, Markdown ZIP export, iCal import
- **i18n** -- English and Japanese

## Quick Start

```bash
cp .env.example .env
docker compose up -d
```

Open `http://localhost:8080` and create your account. The dev compose file sets `DEV_USER=dev@local` to skip registration.

### Production (with Traefik)

```bash
# Edit .env with real credentials and your domain
docker compose -f docker-compose.prod.yml up -d
```

### Native (without Docker)

```bash
# Prerequisites: Go 1.25+, Node.js 22+, PostgreSQL 16+

# Build frontend
cd web/frontend && npm ci && npm run build && cd ../..

# Build server
go build -o bin/server ./cmd/server

# Run migrations
psql $DATABASE_URL -f sql/migrations/0001_initial.up.sql
psql $DATABASE_URL -f sql/migrations/0002_auth_password.up.sql
psql $DATABASE_URL -f sql/migrations/0003_recurring_events.up.sql
psql $DATABASE_URL -f sql/migrations/0004_tool_category.up.sql
psql $DATABASE_URL -f sql/migrations/0005_trgm_search.up.sql

# Start
DATABASE_URL="postgres://..." SESSION_SECRET="..." ./bin/server
```

## Configuration

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | Yes | -- | PostgreSQL connection string |
| `SESSION_SECRET` | Yes | dev fallback | Session signing key |
| `PORT` | No | `8080` | Server port |
| `DEV_USER` | No | -- | Auto-login as this email (dev only) |
| `OPEN_REGISTRATION` | No | -- | Set `true` to allow anyone to register (for Cloud) |

### Docker Compose (prod) variables

| Variable | Description |
|---|---|
| `POSTGRES_PASSWORD` | PostgreSQL password |
| `KONBU_DOMAIN` | Domain for Traefik TLS routing |

## CLI

The CLI is a standalone client that connects to a remote konbu server via API. Server code is not included in the CLI binary.

```bash
go install github.com/krtw00/konbu/cmd/konbu@latest
```

### Setup

```bash
# Set environment variables (recommended: add to ~/.zshrc or ~/.bashrc)
export KONBU_API=https://konbu.example.com
export KONBU_API_KEY=your-api-key

# Or pass as flags
konbu --api https://... --api-key your-key memo list
```

Generate an API key in **Settings > Security** on the web UI.

### Commands

All commands support `--json` flag for machine-readable output.

```
konbu memo list                        # List memos
konbu memo show <id>                   # Show memo content
konbu memo add "title" -c "content"    # Create memo (-c - for stdin)
konbu memo edit <id> --title "new"     # Update memo
konbu memo rm <id>                     # Delete memo

konbu todo list                        # List todos
konbu todo show <id>                   # Show todo details
konbu todo add "task" -t "tag1,tag2"   # Create todo
konbu todo add "task" -d 2025-04-01    # Create with due date
konbu todo edit <id> --desc "notes"    # Update todo
konbu todo done <id>                   # Mark done
konbu todo reopen <id>                 # Reopen
konbu todo rm <id>                     # Delete

konbu event list                       # List events
konbu event show <id>                  # Show event details
konbu event add "title" -s <RFC3339>   # Create event
konbu event edit <id> --title "new"    # Update event
konbu event rm <id>                    # Delete

konbu tool list                        # List tools
konbu tool add "name" "https://..."    # Add tool
konbu tool edit <id> --category "Dev"  # Update tool
konbu tool rm <id>                     # Delete

konbu tag list                         # List tags
konbu tag rm <id>                      # Delete tag

konbu search "query"                   # Cross-search

konbu api-key list                     # List API keys
konbu api-key create "key-name"        # Create API key
konbu api-key rm <id>                  # Delete API key

konbu export json -o backup.json       # Export as JSON
konbu export markdown -o backup.zip    # Export as Markdown ZIP
konbu import ical calendar.ics         # Import iCal file
```

Short IDs (first 8 chars) can be used in place of full UUIDs.

## API

Base path: `/api/v1`

| Resource | Endpoints |
|---|---|
| Auth | `POST /auth/register`, `POST /auth/login`, `POST /auth/logout` |
| User | `GET/PUT /auth/me`, `GET/PUT /auth/settings`, `POST /auth/change-password` |
| API Keys | `GET/POST /api-keys`, `DELETE /api-keys/:id` |
| Memos | `GET/POST /memos`, `GET/PUT/DELETE /memos/:id` |
| ToDos | `GET/POST /todos`, `GET/PUT/DELETE /todos/:id`, `PATCH /todos/:id/done`, `PATCH /todos/:id/reopen` |
| Events | `GET/POST /events`, `GET/PUT/DELETE /events/:id` |
| Tools | `GET/POST /tools`, `PUT/DELETE /tools/:id` |
| Tags | `GET/POST /tags`, `PUT/DELETE /tags/:id` |
| Search | `GET /search?q=...` |
| Export | `GET /export/json`, `GET /export/markdown` |
| Import | `POST /import/ical` |

## Development

```bash
# Start PostgreSQL
docker compose up -d postgres

# Frontend dev server
cd web/frontend && npm run dev

# Run server
DEV_USER=dev@local go run ./cmd/server

# Build CLI
go build -o bin/konbu ./cmd/konbu

# Run tests
go test ./...
```

### Project Structure

```
cmd/
  server/       # API server
  konbu/        # CLI client
internal/
  handler/      # HTTP handlers
  service/      # Business logic
  repository/   # DB access (sqlc)
  middleware/    # Auth, logging
  client/       # API client (used by CLI)
web/frontend/   # React + Vite SPA
sql/            # Schema and migrations
docker/         # Dockerfile
```

## Roadmap

- Reminder notifications (email, browser)
- Mobile UI improvements
- CI test coverage
- AI chat enhancements (context improvements, new model support)

## Sponsors

If you find konbu useful, consider [sponsoring](https://github.com/sponsors/krtw00) the project.

## License

[MIT](LICENSE)
