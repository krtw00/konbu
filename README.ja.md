<p align="center">
  <img src="web/static/favicon.svg" width="64" height="64" alt="konbu">
</p>

<h1 align="center">konbu</h1>

<p align="center">
  <strong>AI執事つきの、自分だけのデジタルシステム手帳。</strong><br>
  CLI · MCP · セルフホスト · AIネイティブ
</p>

<p align="center">
  メモ・ToDo・予定・テーブルを Go バイナリ 1 個の手帳に綴じ、 1 つの場所から横断検索する。<br>
  Claude、Cursor、その他の MCP クライアントから直接操作可能。
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="MIT License"></a>
  <a href="https://github.com/krtw00/konbu/actions"><img src="https://github.com/krtw00/konbu/actions/workflows/deploy.yml/badge.svg" alt="Deploy"></a>
</p>

<p align="center">日本語 | <a href="README.md">English</a></p>

<p align="center"><img src="docs/screenshot.png" width="800" alt="konbu スクリーンショット"></p>

<p align="center"><img src="docs/demo.gif" width="800" alt="AIチャットデモ — 自然言語で予定とToDoを管理"></p>

---

## 使ってみる

- **クラウド版** -- [konbu-cloud.codenica.dev](https://konbu-cloud.codenica.dev) ですぐに使えます（無料・登録のみ）
- **セルフホスト** -- Docker で自分のサーバーに構築（下記参照）

## konbuとは

konbu は **自分だけのデジタルシステム手帳** -- 紙のシステム手帳をデジタル化したものです。 セルフホスト可能な Go バイナリ 1 個に、 メモ・ToDo・予定・構造化テーブルを 1 冊に綴じ、 AI 執事に世話させ、 1 つの検索インターフェイスで「ここ見れば全部書いてある」 を実現します。 他人とは一切共有しない自分専用の一冊で、 Notion + Todoist + Calendar の代替ではなく、 **4 つのアプリを行き来して 1 つのことを探す行為そのもの**の代替です。

何が違うか:

- **MCP サーバー + CLI クライアント** -- AI 執事 (Claude / Cursor / 任意の MCP クライアント) と shell / スクリプト、 2 経路から手帳を操作できる。
- **横断全文検索** -- メモ・ToDo・予定・構造化テーブルを 1 クエリで横断検索。 これは副次機能ではなく**コア UX**。
- **構造化テーブル** (= table-memo、 開発予定) -- 血圧記録・家計簿・在庫など、 Markdown では扱えない表形式データを行 × 列で管理。
- **BYOK AI チャット** -- 自分の OpenAI / Anthropic API キーを使う、 または同梱の無料枠を使う。
- **個人専用設計** -- 自分だけの一冊。 外 (Google カレンダー) からは取り込むが、 外には一切漏らさない。
- **セルフホスト可能** -- Go バイナリ 1 個、 Docker compose、 またはホスト版。

「予定・メモ・タスクがあちこちのアプリに散らばってる状態」を、ひとつのGoバイナリで終わらせます。

## 機能

- **横断全文検索** -- メモ・ToDo・予定・構造化テーブルを 1 クエリで横断検索 (コア UX)
- **CLI & MCPサーバー** -- CLIクライアントとMCPサーバーを同梱。Claude/Cursor等のAIエージェントから直接操作可能
- **AIエージェントチャット** -- 「明日の予定を教えて」「買い物リストをToDoに追加」を自然言語で。BYOK対応、無料枠あり
- **メモ** -- Markdown対応、ライブプレビュー、タグ管理
- **ToDo** -- インライン作成、期限設定、タグフィルタ、ノート付き
- **カレンダー** -- 月表示、予定の作成・編集、iCalインポート（個人専用・owner-only）
- **構造化テーブル** (= table-memo、 開発予定) -- 血圧記録・家計簿・在庫など、 表形式データを行 × 列で管理
- **エクスポート/インポート** -- JSON・Markdown ZIP出力、iCal取り込み
- **多言語対応** -- 日本語・英語

## クイックスタート

```bash
cp .env.example .env
docker compose up -d
```

`http://localhost:8080` を開いてアカウントを作成します。開発用composeでは `DEV_USER=dev@local` が設定されており、登録なしで利用できます。

### 本番環境（Traefik連携）

```bash
# .env にドメインとパスワードを設定して起動
docker compose -f docker-compose.prod.yml up -d
```

### ネイティブ（Docker不要）

```bash
# 前提: Go 1.25+, Node.js 22+, PostgreSQL 16+

# フロントエンドビルド
cd web/frontend && npm ci && npm run build && cd ../..

# サーバービルド
go build -o bin/server ./cmd/server

# 起動（全マイグレーションは起動時に自動適用）
DATABASE_URL="postgres://..." SESSION_SECRET="..." ./bin/server
```

## 設定

| 変数 | 必須 | デフォルト | 説明 |
|---|---|---|---|
| `DATABASE_URL` | Yes | -- | PostgreSQL接続文字列 |
| `SESSION_SECRET` | Yes | 開発用フォールバック | セッション署名キー |
| `PORT` | No | `8080` | サーバーポート |
| `DEV_USER` | No | -- | 開発用自動ログイン（メール形式） |
| `OPEN_REGISTRATION` | No | -- | `true` で誰でもアカウント作成可能（Cloud版向け） |
| `BASE_URL` | No | -- | OAuth コールバックに使う公開URL |
| `GOOGLE_CLIENT_ID` | No | -- | Google OAuth ログイン有効化 |
| `GOOGLE_CLIENT_SECRET` | No | -- | Google OAuth ログイン有効化 |
| `WEBHOOK_SECRET` | No | -- | GitHub Sponsors Webhook シークレット |
| `GITHUB_FEEDBACK_TOKEN` | No | -- | 匿名化したフィードバックを GitHub issue 化するためのトークン |
| `GITHUB_FEEDBACK_REPO` | No | -- | フィードバックを送るリポジトリ。例: `krtw00/konbu` |
| `GITHUB_FEEDBACK_LABELS` | No | -- | issue に付けるカンマ区切りラベル |
| `AI_ENCRYPTION_KEY` | No | -- | BYOK 用 AI キー暗号化に使う64桁hex |
| `DEFAULT_AI_PROVIDER` | No | `openai` | サーバー提供の無料枠 AI プロバイダ |
| `DEFAULT_AI_API_KEY` | No | -- | サーバー提供の無料枠 AI キー |
| `DEFAULT_AI_ENDPOINT` | No | -- | 無料枠プロバイダ endpoint 上書き |
| `DEFAULT_AI_MODEL` | No | -- | 無料枠モデル上書き |
| `R2_ACCESS_KEY_ID` | No | -- | 添付ファイル保存用資格情報 |
| `R2_SECRET_ACCESS_KEY` | No | -- | 添付ファイル保存用資格情報 |
| `R2_ENDPOINT` | No | Cloudflare R2 既定値 | 添付ファイル保存先 endpoint |
| `R2_BUCKET` | No | `konbu-attachments` | 添付ファイル保存先バケット |
| `R2_PUBLIC_URL` | No | -- | 添付ファイル公開ベースURL（任意） |
| `SMTP_HOST` | No | -- | リマインダーメール用 SMTP リレーホスト (例 `smtp-relay.brevo.com`)。 `SMTP_*` 5 項目全てが揃っている時のみ通知機能が起動する |
| `SMTP_PORT` | No | -- | SMTP リレーポート (STARTTLS なら通常 `587`) |
| `SMTP_USERNAME` | No | -- | SMTP リレーログイン名 |
| `SMTP_PASSWORD` | No | -- | SMTP リレーパスワード / API key |
| `SMTP_FROM` | No | -- | リマインダー送信元アドレス |
| `NOTIFICATION_TICK_INTERVAL` | No | `1m` | 通知 sweep の周期 (Go duration、 例 `30s` / `2m`) |

### リマインダー通知

`SMTP_*` env が全て揃っている時、 サーバーは予定 (event) 開始前 と ToDo 期日到来時に email リマインダーを送る in-process loop を起動する。 ユーザーごとに **Settings** (= `user_settings.notifications.enabled = true`) で opt-in、 送信先 email / lead 時間 / 通知時刻 / timezone を個別に上書きできる。

通知は **server 専用機能** (= API サーバープロセス内で動作、 PostgreSQL 必須)。 MCP `--standalone` モード (SQLite) ではリマインダーは送られない。

### Docker Compose（本番用）変数

| 変数 | 説明 |
|---|---|
| `POSTGRES_PASSWORD` | PostgreSQLパスワード |
| `KONBU_DOMAIN` | Traefik TLSルーティング用ドメイン |

## CLI

CLIはリモートのkonbuサーバーにAPI経由で接続するスタンドアロンクライアントです。サーバーのコードはCLIバイナリに含まれません。

```bash
go install github.com/krtw00/konbu/cmd/konbu@latest
```

### セットアップ

```bash
# 環境変数を設定（~/.zshrc 等に追記推奨）
export KONBU_API=https://konbu.example.com
export KONBU_API_KEY=your-api-key

# フラグでも指定可能
konbu --api https://... --api-key your-key memo list
```

APIキーはWeb UIの **設定 > セキュリティ** で発行できます。

### コマンド一覧

全コマンドで `--json` フラグを使うと機械可読なJSON出力になります。

```
konbu memo list                        # メモ一覧
konbu memo show <id>                   # メモ内容を表示
konbu memo add "タイトル" -c "内容"     # メモ作成（-c - で標準入力）
konbu memo edit <id> --title "新名"    # メモ更新
konbu memo rm <id>                     # メモ削除

konbu todo list                        # ToDo一覧
konbu todo show <id>                   # ToDo詳細
konbu todo add "タスク" -t "tag1,tag2" # ToDo作成
konbu todo add "タスク" -d 2025-04-01  # 期限付きで作成
konbu todo edit <id> --desc "メモ"     # ToDo更新
konbu todo done <id>                   # 完了にする
konbu todo reopen <id>                 # 未完了に戻す
konbu todo rm <id>                     # 削除

konbu event list                       # 予定一覧
konbu event show <id>                  # 予定詳細
konbu event add "タイトル" -s <RFC3339> # 予定作成
konbu event edit <id> --title "新名"   # 予定更新
konbu event rm <id>                    # 削除

konbu tag list                         # タグ一覧
konbu tag rm <id>                      # タグ削除

konbu search "検索語"                  # 横断検索

konbu api-key list                     # APIキー一覧
konbu api-key create "キー名"          # APIキー作成
konbu api-key rm <id>                  # APIキー削除

konbu export json -o backup.json       # JSONエクスポート
konbu export markdown -o backup.zip    # Markdown ZIPエクスポート
konbu import ical calendar.ics         # iCalインポート
```

IDは先頭8文字の短縮形で指定できます。

## MCPサーバー

konbu は MCP（Model Context Protocol）サーバーを内蔵しており、2つのモードで動かせます。用途に応じて選んでください。

### スタンドアロンモード（SQLite、サーバー不要）

ローカルの MCP バックエンドとして使いたいだけなら、CLI をインストールして `--standalone` で起動するだけ。PostgreSQL も Web サーバーも API キーも不要 — すべてローカル SQLite ファイルに保存されます。

```bash
go install github.com/krtw00/konbu/cmd/konbu@latest
konbu mcp --standalone
```

データは既定で `~/.konbu/konbu.db` に保存されます。`--db /path/to/db.sqlite` で上書き可能。

**Claude Desktop**（macOS は `~/Library/Application Support/Claude/claude_desktop_config.json`、Windows は `%APPDATA%\Claude\claude_desktop_config.json`）:

```json
{
  "mcpServers": {
    "konbu": {
      "command": "konbu",
      "args": ["mcp", "--standalone"]
    }
  }
}
```

**Cursor** も同じ設定を `~/.cursor/mcp.json`（または設定 UI）に追加すれば動きます。

#### Docker

マルチアーキ（`linux/amd64`・`linux/arm64`）の公式イメージを GitHub Container Registry で配布しています。ビルド不要でそのまま pull できます:

```bash
docker pull ghcr.io/krtw00/konbu-mcp:latest
```

固定バージョンを使いたい場合はリリースタグを指定してください（例: `docker pull ghcr.io/krtw00/konbu-mcp:v0.2.0`）。

MCP クライアントからは名前付きボリュームを使って起動します:

```json
{
  "mcpServers": {
    "konbu": {
      "command": "docker",
      "args": ["run", "--rm", "-i", "-v", "konbu-data:/data", "ghcr.io/krtw00/konbu-mcp:latest"]
    }
  }
}
```

ソースからビルドしたい場合は、リポジトリのルートで `docker build -f docker/Dockerfile.mcp -t konbu-mcp .` を実行すれば同じイメージが手元で作れます（CGO 不要 / distroless static / 約 22 MB）。

スタンドアロンモードはメモ / ToDo / 予定 の CRUD と横断検索を提供します。タグ・添付ファイル・AI チャットはサーバー専用機能なので、それらが必要な場合は下記の接続モードを使ってください。

### 接続モード（konbu サーバーに接続）

セルフホスト中の konbu サーバーまたは [konbu Cloud](https://konbu-cloud.codenica.dev) を MCP の裏側にしたい場合は HTTP 経由で接続します。タグ・添付・AI チャットを含む全機能が使えます。

1. `konbu` CLI バイナリをインストール（上の [CLI](#cli) セクション参照）
2. Web UI の **設定 > セキュリティ** で API キーを発行
3. お使いの MCP クライアント設定に konbu を追加:

```json
{
  "mcpServers": {
    "konbu": {
      "command": "konbu",
      "args": ["mcp"],
      "env": {
        "KONBU_API": "http://localhost:8080",
        "KONBU_API_KEY": "your-api-key"
      }
    }
  }
}
```

### 使用例

MCP クライアントを再起動すると、自然言語で konbu を操作できます:

- 「明日の予定を教えて」
- 「金曜13時に歯医者の予定を入れて」
- 「『買い物』タグで牛乳を買うのToDoを追加」
- 「先週の『会議』タグのメモを見せて」
- 「『PRレビュー』のToDoを完了にして」

## API

ベースパス: `/api/v1`

| リソース | エンドポイント |
|---|---|
| 認証 | `POST /auth/register`, `POST /auth/login`, `POST /auth/logout`, `GET /auth/setup-status`, `GET /auth/providers`, `GET /auth/google/login`, `GET /auth/google/callback` |
| ユーザー | `GET/PUT /auth/me`, `GET/PUT /auth/settings`, `POST /auth/change-password`, `POST /auth/delete-account` |
| APIキー | `GET/POST /api-keys`, `DELETE /api-keys/:id` |
| メモ | `GET/POST /memos`, `GET/PUT/DELETE /memos/:id`, `GET/POST /memos/:id/rows`, `POST /memos/:id/rows/batch`, `GET /memos/:id/rows/export`, `PUT/DELETE /memos/:id/rows/:rowId` |
| ToDo | `GET/POST /todos`, `GET/PUT/DELETE /todos/:id`, `PATCH /todos/:id/done`, `PATCH /todos/:id/reopen` |
| 予定 | `GET/POST /events`, `GET/PUT/DELETE /events/:id` |
| カレンダー | `GET/POST /calendars`, `GET/PUT/DELETE /calendars/:id`（owner-only） |
| タグ | `GET/POST /tags`, `PUT/DELETE /tags/:id` |
| 検索 | `GET /search?q=...` |
| チャット | `GET/POST /chat/sessions`, `GET/PUT/DELETE /chat/sessions/:id`, `POST /chat/sessions/:id/messages`, `GET/PUT /chat/config` |
| 添付ファイル | `POST /attachments`, `GET /attachments/*` |
| エクスポート | `GET /export/json`, `GET /export/markdown` |
| インポート | `POST /import/ical` |

## 開発

```bash
# PostgreSQL起動
docker compose up -d postgres

# フロントエンド開発サーバー
cd web/frontend && npm run dev

# サーバー起動
DEV_USER=dev@local go run ./cmd/server

# CLIビルド
go build -o bin/konbu ./cmd/konbu

# テスト
go test ./...
```

### ディレクトリ構成

```
cmd/
  server/       # APIサーバー
  konbu/        # CLIクライアント
internal/
  handler/      # HTTPハンドラ
  service/      # ビジネスロジック
  repository/   # DBアクセス (sqlc)
  middleware/   # 認証・ログ
  client/       # APIクライアント（CLI用）
  mcp/          # MCPサーバー
web/frontend/   # React + Vite SPA
sql/            # スキーマ・マイグレーション
docker/         # Dockerfile
```

## Roadmap

- ブラウザ push 通知 (= メール通知は `SMTP_*` env が揃えば既に利用可能)
- スマホUI改善
- CI テスト追加
- AIチャット強化（コンテキスト改善、新モデル対応）

## スポンサー

konbuが役に立ったら、[スポンサー](https://github.com/sponsors/krtw00)でプロジェクトを支援できます。

## ライセンス

[MIT](LICENSE)
