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
```

## Contributing

Contributions are welcome! Whether it's reporting a bug, requesting a feature, or submitting a Pull Request, please feel free to open an issue or reach out. For code contributions, please ensure your code follows the existing style and includes tests where applicable.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.