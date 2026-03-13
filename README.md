<p align="center">
  <img src="web/static/favicon.svg" width="64" height="64" alt="konbu">
</p>

<h1 align="center">konbu</h1>

<p align="center">Self-hosted personal tool hub: memos, todos, calendar, and tool launcher in one place.</p>

---

## Features

- **Memos** -- Markdown and table-type notes with tagging, full-screen CodeMirror 6 editor, and live preview
- **ToDo** -- Inline task creation, due dates, tag filtering, and detail panel
- **Calendar** -- Monthly view with event creation and editing
- **Tools** -- Bookmark launcher with automatic favicon fetching
- **Cross-search** -- Full-text search across memos, todos, and events (pg_bigm)
- **CLI** -- Manage memos, todos, and tools from the terminal
- **Themes** -- 7 built-in color themes (Konbu, Notion, Solarized, Catppuccin Latte/Mocha, Nord, Linear)

## Quick Start

```bash
cp .env.example .env
docker compose up -d
```

Open `http://localhost:8080`. The dev compose file sets `DEV_USER=dev@local` to bypass authentication.

### Production (with Traefik)

```bash
# Edit .env with real credentials and your domain
docker compose -f docker-compose.prod.yml up -d
```

## Configuration

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | Yes | -- | PostgreSQL connection string |
| `PORT` | No | `8080` | Server port |
| `KONBU_USER` | Prod | -- | Login username |
| `KONBU_PASS` | Prod | -- | Login password |
| `SESSION_SECRET` | Prod | -- | Session signing key |
| `ADMIN_EMAIL` | No | -- | Admin user email |
| `ALLOWED_EMAILS` | No | `*` | Comma-separated allowed emails (`*` = all) |
| `DEV_USER` | No | -- | Dev mode: bypass auth with this email |
| `POSTGRES_PASSWORD` | Prod | -- | PostgreSQL password (prod compose) |
| `KONBU_DOMAIN` | Prod | -- | Domain for Traefik routing (prod compose) |

## API

Base path: `/api/v1`

| Resource | Endpoints |
|---|---|
| Memos | `GET/POST /memos`, `GET/PUT/DELETE /memos/:id` |
| ToDos | `GET/POST /todos`, `GET/PUT/DELETE /todos/:id`, `PATCH /todos/:id/done` |
| Events | `GET/POST /events`, `GET/PUT/DELETE /events/:id` |
| Tools | `GET/POST /tools`, `PUT/DELETE /tools/:id` |
| Tags | `GET/POST /tags`, `PUT/DELETE /tags/:id` |
| Search | `GET /search?q=...` |
| Auth | `GET/PUT /auth/me`, `GET/POST/DELETE /api-keys` |

See [docs/api.md](docs/api.md) for full specification.

## CLI

```bash
go build -o bin/konbu ./cmd/konbu

konbu memo list
konbu memo add "title" -c "content"
konbu todo list
konbu todo add "task name"
konbu todo done <id>
konbu tool list
konbu tool add "name" "https://..."
```

Set `KONBU_API` or use `--api` flag to point to your server.

## Development

```bash
# Start PostgreSQL
docker compose up -d postgres

# Run server
go run ./cmd/server

# Generate repository code from SQL
sqlc generate

# Run tests
go test ./...
```

### Project Structure

```
cmd/
  server/       # API server
  konbu/        # CLI
internal/
  handler/      # HTTP handlers
  service/      # Business logic
  repository/   # DB access (sqlc)
  middleware/    # Auth, logging
web/static/     # Frontend (HTML/CSS/JS)
sql/            # Schema and migrations
```

## License

[MIT](LICENSE)
