# Contributing to konbu

[日本語](#日本語) | [English](#english)

---

## English

Thank you for your interest in contributing to konbu!

### What is konbu?

konbu is a personal workspace that provides memos, todos, calendar, AI chat, and bookmark management via REST API + Web UI + CLI.

- **Backend**: Go (chi v5, sqlc, PostgreSQL)
- **Frontend**: React 19 + TypeScript (Vite, shadcn/ui, Zustand)
- **CLI / MCP**: cobra, stdio

### Prerequisites

| Tool | Version |
|------|---------|
| Go | 1.25+ |
| Node.js | 22+ |
| Docker / Docker Compose | Latest |

### Setup

```bash
# 1. Clone
git clone https://github.com/krtw00/konbu.git
cd konbu

# 2. Start PostgreSQL
docker compose up -d postgres

# 3. Install frontend dependencies
cd web/frontend && npm ci && cd ../..

# 4. Start server (dev mode)
DATABASE_URL="postgres://konbu:konbu@localhost:5432/konbu?sslmode=disable" \
DEV_USER=dev@local \
go run ./cmd/server

# 5. Start frontend dev server (separate terminal)
cd web/frontend && npm run dev
```

Open `http://localhost:5173` in your browser. API requests are proxied to `http://localhost:8080`.

#### All-in-one with Docker

```bash
docker compose up -d
# → http://localhost:8080
```

### Architecture

```
Request → middleware(auth) → handler → service → repository → PostgreSQL
                              ↓
                           Response (JSON)
```

#### Directory structure

```
konbu/
├── cmd/server/          # API server entrypoint
├── cmd/konbu/           # CLI entrypoint
├── internal/
│   ├── handler/         # HTTP handlers (input/output only)
│   ├── service/         # Business logic
│   ├── repository/      # DB access (sqlc + hand-written SQL)
│   ├── model/           # Request/response structs
│   ├── middleware/       # Auth, logging
│   ├── apperror/        # Application error types
│   ├── mcp/             # MCP Server implementation
│   └── client/          # HTTP client for CLI
├── sql/
│   ├── schema.sql       # DDL (source of truth)
│   ├── migrations/      # Migration files (NNNN_description.up/down.sql)
│   └── queries/         # sqlc query files
├── web/frontend/        # React frontend
│   ├── src/pages/       # Page components
│   ├── src/components/  # Shared components (shadcn/ui)
│   ├── src/stores/      # Zustand stores
│   ├── src/lib/         # Utilities
│   ├── src/hooks/       # Custom hooks
│   ├── src/i18n/        # Translation files (en.json, ja.json)
│   └── src/types/       # TypeScript type definitions
├── docs/                # Design documents
└── docker/              # Dockerfile
```

#### Layer responsibilities

| Layer | Does | Does NOT do |
|-------|------|-------------|
| handler | HTTP I/O, validation, response formatting | Business logic, DB access |
| service | Business logic, transaction management | HTTP, SQL |
| repository | DB access only | Business logic |

### Development workflow

#### 1. Open an issue (recommended)

For large changes or new features, discuss the approach in an issue first. Small bug fixes and typos can go straight to a PR.

For collaborator-facing issue slicing and project workflow, see [docs/plans/27-collaboration-workflow.md](docs/plans/27-collaboration-workflow.md).

Recommended conventions:

- Use one `type:*`, one `area:*`, and one `prio:*` label per issue
- Track progress in GitHub Project `Status` instead of status labels
- Split large initiatives into one parent issue plus `S`/`M` sized task issues
- Keep one PR mapped to one issue whenever possible

#### 2. Create a branch

```bash
git checkout -b feat/your-feature   # New feature
git checkout -b fix/your-bugfix     # Bug fix
```

#### 3. Make your changes

- Run Go tests: `go test ./...`
- Type-check frontend: `cd web/frontend && npx tsc --noEmit`
- Build frontend: `cd web/frontend && npm run build`

#### 4. Commit messages

[Conventional Commits](https://www.conventionalcommits.org/) preferred:

```
feat: add tag-based search for memos
fix: calendar event times off by 9 hours
refactor: extract cache invalidation logic
docs: update API documentation
```

#### 5. Open a Pull Request

- One PR per change
- Describe **why** the change is needed
- Include screenshots for UI changes

### Coding conventions

#### Go

- Package names: lowercase single word (`handler`, `service`)
- File names: snake_case (`memo_handler.go`)
- Structs: PascalCase (`CreateMemoRequest`)
- Errors: use types from `internal/apperror`
- Layers depend on each other via interfaces

#### Frontend

- TypeScript required (avoid `any`)
- Components: shadcn/ui based
- State management: Zustand (minimal global state, prefer local state)
- i18n: all user-facing text in `en.json` / `ja.json`
- API calls: centralized in `src/lib/api.ts`
- After data changes: call `invalidateCache()` for related caches

#### SQL

- Table names: snake_case plural (`memos`, `calendar_events`)
- Column names: snake_case (`user_id`, `created_at`)
- Migrations: `NNNN_description.up.sql` / `.down.sql`
- Soft delete: `deleted_at` column, all SELECTs include `WHERE deleted_at IS NULL`

#### API

- Paths: kebab-case plural (`/api/v1/api-keys`)
- JSON fields: snake_case (`created_at`)
- Error response: `{"error": {"code": "xxx", "message": "xxx"}}`

### Common development tasks

#### Adding a DB migration

```bash
ls sql/migrations/              # Check next number
# Create: sql/migrations/NNNN_description.up.sql
# Create: sql/migrations/NNNN_description.down.sql
# Update: sql/schema.sql and docs/schema.sql
```

Migrations are auto-applied on server startup.

#### Generating sqlc code

```bash
sqlc generate    # After editing sql/queries/*.sql
```

#### Adding a new page

1. Create `web/frontend/src/pages/XxxPage.tsx`
2. Add to `Page` type in `src/stores/app.ts`
3. Add routing in `src/App.tsx`
4. Add nav item in `Sidebar.tsx` and `MobileHeader.tsx`
5. Add translations to `en.json` and `ja.json`

#### Adding a new API endpoint

1. Add query in `sql/queries/` → `sqlc generate`
2. Add repository method in `internal/repository/`
3. Add service method in `internal/service/`
4. Add handler in `internal/handler/`
5. Add route in `cmd/server/main.go`
6. Add client method in `web/frontend/src/lib/api.ts`

### Environment variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | — | PostgreSQL connection string |
| `SESSION_SECRET` | Yes | `konbu-dev-secret-change-me` | Session signing key |
| `PORT` | No | `8080` | API server port |
| `DEV_USER` | No | — | Auto-login user for development (email format) |

### Testing

```bash
go test ./...                              # All Go tests
go test ./internal/handler/                # Specific package
cd web/frontend && npx tsc --noEmit        # Frontend type check
```

### License

All contributions are released under the [MIT License](LICENSE).

---

## 日本語

konbu へのコントリビュートに興味を持っていただきありがとうございます！

### konbu とは

konbu はパーソナルワークスペースです。メモ・ToDo・カレンダー・AIチャット・ブックマーク管理を REST API + Web UI + CLI で提供しています。

- **バックエンド**: Go (chi v5, sqlc, PostgreSQL)
- **フロントエンド**: React 19 + TypeScript (Vite, shadcn/ui, Zustand)
- **CLI / MCP**: cobra, stdio

### 前提条件

| ツール | バージョン |
|--------|-----------|
| Go | 1.25+ |
| Node.js | 22+ |
| Docker / Docker Compose | 最新 |

### セットアップ

```bash
# 1. クローン
git clone https://github.com/krtw00/konbu.git
cd konbu

# 2. PostgreSQL 起動
docker compose up -d postgres

# 3. フロントエンド依存インストール
cd web/frontend && npm ci && cd ../..

# 4. サーバー起動（開発モード）
DATABASE_URL="postgres://konbu:konbu@localhost:5432/konbu?sslmode=disable" \
DEV_USER=dev@local \
go run ./cmd/server

# 5. フロントエンド開発サーバー（別ターミナル）
cd web/frontend && npm run dev
```

ブラウザで `http://localhost:5173` を開いてください。API は `http://localhost:8080` にプロキシされます。

#### Docker でまとめて起動

```bash
docker compose up -d
# → http://localhost:8080
```

### アーキテクチャ

```
Request → middleware(認証) → handler → service → repository → PostgreSQL
                              ↓
                           Response (JSON)
```

#### ディレクトリ構成

```
konbu/
├── cmd/server/          # API サーバーエントリポイント
├── cmd/konbu/           # CLI エントリポイント
├── internal/
│   ├── handler/         # HTTP ハンドラ（入出力のみ）
│   ├── service/         # ビジネスロジック
│   ├── repository/      # DB アクセス（sqlc 生成 + 手書き SQL）
│   ├── model/           # リクエスト・レスポンス構造体
│   ├── middleware/       # 認証・ログ
│   ├── apperror/        # アプリケーションエラー型
│   ├── mcp/             # MCP Server 実装
│   └── client/          # CLI 用 HTTP クライアント
├── sql/
│   ├── schema.sql       # DDL 本体
│   ├── migrations/      # マイグレーションファイル
│   └── queries/         # sqlc 用クエリ
├── web/frontend/        # React フロントエンド
│   ├── src/pages/       # ページコンポーネント
│   ├── src/components/  # 共通コンポーネント（shadcn/ui）
│   ├── src/stores/      # Zustand ストア
│   ├── src/lib/         # ユーティリティ
│   ├── src/hooks/       # カスタムフック
│   ├── src/i18n/        # 翻訳ファイル（en.json, ja.json）
│   └── src/types/       # TypeScript 型定義
├── docs/                # 設計ドキュメント
└── docker/              # Dockerfile
```

#### レイヤーの責務

| レイヤー | やること | やらないこと |
|---------|---------|-------------|
| handler | HTTP の入出力、バリデーション、レスポンス整形 | ビジネスロジック、DB アクセス |
| service | ビジネスロジック、トランザクション管理 | HTTP、SQL |
| repository | DB アクセスのみ | ビジネスロジック |

### 開発フロー

#### 1. Issue で相談（推奨）

大きな変更や新機能は、先に Issue で方針を相談してください。小さなバグ修正や typo 修正は直接 PR で OK です。

コラボレーター向けの Issue 分割方針と Project 運用は [docs/plans/27-collaboration-workflow.md](docs/plans/27-collaboration-workflow.md) を参照してください。

推奨ルール:

- 各 Issue には `type:*`、`area:*`、`prio:*` をそれぞれ 1 つずつ付ける
- 進捗管理は status ラベルではなく GitHub Project の `Status` フィールドで行う
- 大きいテーマは親 Issue 1 つと `S` / `M` サイズの task issue に分割する
- 可能な限り 1 PR = 1 Issue に対応させる

#### 2. ブランチを切る

```bash
git checkout -b feat/your-feature   # 機能追加
git checkout -b fix/your-bugfix     # バグ修正
```

#### 3. 変更を加える

- Go テスト確認: `go test ./...`
- フロントエンド型チェック: `cd web/frontend && npx tsc --noEmit`
- ビルド確認: `cd web/frontend && npm run build`

#### 4. コミットメッセージ

[Conventional Commits](https://www.conventionalcommits.org/) を推奨:

```
feat: add tag-based search for memos
fix: calendar event times off by 9 hours
refactor: extract cache invalidation logic
docs: update API documentation
```

#### 5. Pull Request を作成

- 1つの PR で 1つの変更
- 変更の目的（なぜ）を書く
- UI 変更の場合はスクリーンショット添付

### コーディング規約

#### Go

- パッケージ名: 小文字単一語 (`handler`, `service`)
- ファイル名: スネークケース (`memo_handler.go`)
- 構造体: パスカルケース (`CreateMemoRequest`)
- エラー: `internal/apperror` の型を使う
- レイヤー間はインターフェースで依存

#### フロントエンド

- TypeScript 必須（`any` は避ける）
- コンポーネント: shadcn/ui ベース
- 状態管理: Zustand（グローバル最小限、ローカル state 優先）
- i18n: 全ユーザー向けテキストは `en.json` / `ja.json` に定義
- API 呼び出し: `src/lib/api.ts` に集約
- データ変更後は `invalidateCache()` で関連キャッシュを無効化

#### SQL

- テーブル名: スネークケース複数形 (`memos`, `calendar_events`)
- カラム名: スネークケース (`user_id`, `created_at`)
- マイグレーション: `NNNN_description.up.sql` / `.down.sql`
- 論理削除: `deleted_at` カラム、全 SELECT に `WHERE deleted_at IS NULL`

#### API

- パス: ケバブケース複数形 (`/api/v1/api-keys`)
- JSON フィールド: スネークケース (`created_at`)
- エラーレスポンス: `{"error": {"code": "xxx", "message": "xxx"}}`

### よくある開発タスク

#### DB マイグレーション追加

```bash
ls sql/migrations/              # 次の番号を確認
# sql/migrations/NNNN_description.up.sql を作成
# sql/migrations/NNNN_description.down.sql を作成
# sql/schema.sql と docs/schema.sql にも反映
```

サーバー起動時に自動適用されます。

#### sqlc でコード生成

```bash
sqlc generate    # sql/queries/*.sql 編集後
```

#### 新しいページを追加

1. `web/frontend/src/pages/XxxPage.tsx` を作成
2. `src/stores/app.ts` の `Page` 型に追加
3. `src/App.tsx` にルーティング追加
4. `Sidebar.tsx` と `MobileHeader.tsx` にナビ追加
5. `en.json` と `ja.json` に翻訳追加

#### 新しい API エンドポイント追加

1. `sql/queries/` にクエリ追加 → `sqlc generate`
2. `internal/repository/` にリポジトリメソッド追加
3. `internal/service/` にサービスメソッド追加
4. `internal/handler/` にハンドラ追加
5. `cmd/server/main.go` にルーティング追加
6. `web/frontend/src/lib/api.ts` にクライアントメソッド追加

### 環境変数

| 変数 | 必須 | デフォルト | 説明 |
|------|------|-----------|------|
| `DATABASE_URL` | Yes | — | PostgreSQL 接続文字列 |
| `SESSION_SECRET` | Yes | `konbu-dev-secret-change-me` | セッション署名キー |
| `PORT` | No | `8080` | API サーバーポート |
| `DEV_USER` | No | — | 開発用自動ログインユーザー |

### テスト

```bash
go test ./...                              # Go 全テスト
go test ./internal/handler/                # 特定パッケージ
cd web/frontend && npx tsc --noEmit        # フロントエンド型チェック
```

### ライセンス

コントリビュートされたコードは [MIT License](LICENSE) の下で公開されます。
