<p align="center">
  <img src="web/static/favicon.svg" width="64" height="64" alt="konbu">
</p>

<h1 align="center">konbu</h1>

<p align="center">An AI-powered digital planner.<br>Keep your schedule at the center, with notes and todos in the same place.</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="MIT License"></a>
  <a href="https://github.com/krtw00/konbu/actions"><img src="https://github.com/krtw00/konbu/actions/workflows/deploy.yml/badge.svg" alt="Deploy"></a>
</p>

<p align="center"><a href="README.ja.md">日本語</a> | English</p>

<p align="center"><img src="docs/screenshot.png" width="800" alt="konbu screenshot"></p>

<p align="center"><img src="docs/demo.gif" width="800" alt="AI chat demo — managing schedule and todos with natural language"></p>

---

## Try It

- **Cloud** -- Use instantly at [konbu-cloud.codenica.dev](https://konbu-cloud.codenica.dev) (free, no setup)
- **Self-hosted** -- Run on your own server with Docker (see below)

## Why konbu?

Planners are good at showing your schedule, but the notes and tasks around it usually end up in separate apps.

Once that context is scattered, you spend more time remembering where you put things than actually using them.

konbu is an AI-powered digital planner that keeps your schedule, notes, todos, and links in the same place. The calendar stays at the center, while the rest of the information remains close enough to act on and easy enough to find later.

It is text-first, but not text-only. Each type of information still has the UI it needs, and AI sits on top as an agent that can check your schedule, search, organize, rewrite, and add things for you.

## Features

- **Cross-search** -- Full-text search across memos, todos, events, and bookmarks. No more "where was that?"
- **AI Agent Chat** -- "Add groceries to my todo" "What's on my schedule tomorrow?" — AI searches, organizes, rewrites, and acts on your workspace (free tier included)
- **Memos** -- Markdown notes with tagging, live preview
- **ToDo** -- Inline task creation with due dates, tags, and notes
- **Calendar** -- Monthly view with event CRUD and iCal import
- **Tools** -- Site launcher with categories and drag-and-drop reordering
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

# Start (runs all SQL migrations automatically on boot)
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
| `BASE_URL` | No | -- | Public app URL used for OAuth callbacks |
| `GOOGLE_CLIENT_ID` | No | -- | Enable Google OAuth login |
| `GOOGLE_CLIENT_SECRET` | No | -- | Enable Google OAuth login |
| `WEBHOOK_SECRET` | No | -- | GitHub Sponsors webhook secret |
| `KOFI_TOKEN` | No | -- | Ko-fi webhook verification token |
| `GITHUB_FEEDBACK_TOKEN` | No | -- | GitHub token used to create anonymized feedback issues |
| `GITHUB_FEEDBACK_REPO` | No | -- | Repository to receive feedback issues, e.g. `krtw00/konbu` |
| `GITHUB_FEEDBACK_LABELS` | No | -- | Comma-separated labels added to forwarded feedback issues |
| `AI_ENCRYPTION_KEY` | No | -- | 64 hex chars used to encrypt BYOK AI keys |
| `DEFAULT_AI_PROVIDER` | No | `openai` | Server-side free-tier AI provider |
| `DEFAULT_AI_API_KEY` | No | -- | Server-side free-tier AI key |
| `DEFAULT_AI_ENDPOINT` | No | -- | Override free-tier provider endpoint |
| `DEFAULT_AI_MODEL` | No | -- | Override free-tier provider model |
| `R2_ACCESS_KEY_ID` | No | -- | Attachment upload credentials |
| `R2_SECRET_ACCESS_KEY` | No | -- | Attachment upload credentials |
| `R2_ENDPOINT` | No | Cloudflare R2 default | Attachment storage endpoint |
| `R2_BUCKET` | No | `konbu-attachments` | Attachment storage bucket |
| `R2_PUBLIC_URL` | No | -- | Optional public base URL for attachments |

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
| Auth | `POST /auth/register`, `POST /auth/login`, `POST /auth/logout`, `GET /auth/setup-status`, `GET /auth/providers`, `GET /auth/google/login`, `GET /auth/google/callback` |
| User | `GET/PUT /auth/me`, `GET/PUT /auth/settings`, `POST /auth/change-password`, `POST /auth/delete-account` |
| API Keys | `GET/POST /api-keys`, `DELETE /api-keys/:id` |
| Memos | `GET/POST /memos`, `GET/PUT/DELETE /memos/:id`, `GET/POST /memos/:id/rows`, `POST /memos/:id/rows/batch`, `GET /memos/:id/rows/export`, `PUT/DELETE /memos/:id/rows/:rowId` |
| ToDos | `GET/POST /todos`, `GET/PUT/DELETE /todos/:id`, `PATCH /todos/:id/done`, `PATCH /todos/:id/reopen` |
| Events | `GET/POST /events`, `GET/PUT/DELETE /events/:id` |
| Calendars | `GET/POST /calendars`, `GET/PUT/DELETE /calendars/:id`, `POST /calendars/join/:token`, share-link and member management, `GET /calendar.ics` |
| Tools | `GET/POST /tools`, `PUT/DELETE /tools/:id` |
| Tags | `GET/POST /tags`, `PUT/DELETE /tags/:id` |
| Search | `GET /search?q=...` |
| Chat | `GET/POST /chat/sessions`, `GET/PUT/DELETE /chat/sessions/:id`, `POST /chat/sessions/:id/messages`, `GET/PUT /chat/config` |
| Attachments | `POST /attachments`, `GET /attachments/*` |
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
  middleware/   # Auth, logging
  client/       # API client (used by CLI)
  mcp/          # MCP server
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
