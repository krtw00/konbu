# konbu — 設計ドキュメント

> 更新: 2026-03-13

---

## コンセプト

**「ブラウザを開いたらここから始まる、自分専用のワークスペース」**

- メモ・ToDo・カレンダー・ツールランチャーを1つのアプリに統合
- セルフホスト前提。自分のサーバーに置いて完全にコントロールできる
- REST APIが唯一のデータアクセス層。Web UI・CLI・外部連携すべて同じAPIを使う

---

## アーキテクチャ

```
┌─────────────┐  ┌─────────────┐
│   Web UI    │  │     CLI     │
│ React + Vite│  │  cobra/Go   │
└──────┬──────┘  └──────┬──────┘
       │                │
       │  HTTP/JSON      │  HTTP/JSON + Bearer Token
       ▼                ▼
┌─────────────────────────────────┐
│         REST API Server         │
│         Go / chi router         │
├─────────────────────────────────┤
│  middleware: Auth, Session, Log │
├─────────────────────────────────┤
│  handler → service → repository │
├─────────────────────────────────┤
│         PostgreSQL 16+          │
│    pg_bigm (日本語全文検索)      │
└─────────────────────────────────┘
```

### レイヤー構成

```
handler   HTTP入出力のみ。バリデーション、レスポンス整形
service   ビジネスロジック。トランザクション管理、タグのupsert
repository  DBアクセス。sqlc生成コード + カスタムクエリ
```

各レイヤーはインターフェース経由で依存し、テスト容易性を確保する。

---

## 技術スタック

| レイヤー | 技術 |
|---------|------|
| バックエンド | Go / chi / sqlc / cobra |
| フロントエンド | React 19 / TypeScript / shadcn/ui / Zustand / Vite |
| エディタ | CodeMirror 6（@uiw/react-codemirror） |
| i18n | i18next / react-i18next（日本語・英語） |
| DB | PostgreSQL 16+ |
| 全文検索 | pg_bigm（2-gram、日本語対応） |
| コンテナ | Docker（マルチステージビルド、scratchベース） |
| 本番デプロイ | Docker Compose or systemd + リバースプロキシ |

---

## データモデル

### テーブル一覧

| テーブル | 説明 |
|---------|------|
| users | ユーザー（メール+パスワード認証） |
| api_keys | APIキー（CLI/外部連携用、Bearer認証） |
| tags | タグ（ユーザーごとの名前空間） |
| memos | メモ（Markdown / テーブル型） |
| memo_tags | メモ↔タグ中間テーブル |
| memo_rows | テーブル型メモの行データ（JSONB） |
| todos | ToDo（status: open/done、期限付き） |
| todo_tags | ToDo↔タグ中間テーブル |
| calendar_events | カレンダー予定（終日/時間指定、繰り返し対応） |
| calendar_event_tags | 予定↔タグ中間テーブル |
| tools | ツールランチャー（URL+favicon、カテゴリ分類） |

全テーブルに `deleted_at` による論理削除を採用。`WHERE deleted_at IS NULL` を基本とする。

### ER図（簡略）

```
users ─┬─< api_keys
       ├─< tags
       ├─< memos ──< memo_tags >── tags
       │     └──< memo_rows
       ├─< todos ──< todo_tags >── tags
       ├─< calendar_events ──< calendar_event_tags >── tags
       └─< tools
```

全リソースは `user_id` でスコープされ、マルチユーザー環境でデータが分離される。

---

## 認証

3つの認証方式を優先順位で評価する:

1. **APIキー認証** — `Authorization: Bearer <key>` ヘッダー。CLIや外部連携で使用。api_keysテーブルのkey_hashと照合
2. **セッション認証** — `konbu_session` Cookie。HMAC-SHA256で署名されたuser_idをペイロードに持つ。30日有効。Web UIで使用
3. **開発モード** — `DEV_USER` 環境変数設定時、そのメールアドレスで自動ログイン（認証スキップ）

初回起動時はユーザーが存在しないため、`/api/v1/auth/setup-status` で未登録状態を検出し、Web UIがアカウント作成画面を表示する。

---

## モジュール構成

### ドメインモジュール

| モジュール | handler | service | 機能 |
|-----------|---------|---------|------|
| auth | auth_handler | auth_service | ユーザー登録/ログイン/設定/APIキー管理 |
| memo | memo_handler | memo_service | メモCRUD、Markdown/テーブル型 |
| todo | todo_handler | todo_service | ToDo CRUD、完了/未完了切り替え |
| event | event_handler | event_service | カレンダー予定CRUD、繰り返し対応 |
| tool | tool_handler | tool_service | ツールCRUD、favicon自動取得、ヘルスチェック |
| tag | tag_handler | tag_service | タグCRUD、暗黙的upsert |
| search | search_handler | search_service | メモ・ToDo・予定の横断全文検索 |
| export | export_handler | export_service | JSON/Markdown ZIPエクスポート |
| import | import_handler | import_service | iCal（RFC 5545）インポート |

### ユーティリティ

| ファイル | 機能 |
|---------|------|
| response.go | 共通レスポンスヘルパー |
| favicon.go | URL→faviconフェッチ |

---

## CLI設計

CLIはリモートAPIクライアントとして動作する。サーバーコードには依存しない。

```
cmd/konbu/main.go   CLI本体（cobraコマンド定義）
internal/client/    HTTPクライアント（APIアクセス層）
```

### 設計方針

- `go install github.com/krtw00/konbu/cmd/konbu@latest` でCLIのみインストール可能
- サーバーの handler/service/repository はCLIバイナリに含まれない
- `--json` フラグで全コマンドが機械可読なJSON出力に対応（AI連携を想定）
- 短縮ID（先頭8文字）でリソースを指定可能
- `KONBU_API` + `KONBU_API_KEY` 環境変数でリモートサーバーに接続

### コマンド体系

```
konbu memo     list | show | add | edit | rm
konbu todo     list | show | add | edit | done | reopen | rm
konbu event    list | show | add | edit | rm
konbu tool     list | add | edit | rm
konbu tag      list | rm
konbu search   <query>
konbu api-key  list | create | rm
konbu export   json | markdown
konbu import   ical
```

---

## フロントエンド

### 構成

```
web/frontend/
  src/
    pages/        ページコンポーネント（ルートごと）
    components/   共通UIコンポーネント（shadcn/ui）
    stores/       状態管理（Zustand）
    lib/          APIクライアント、ユーティリティ
    i18n/         翻訳ファイル（en.json, ja.json）
```

### ページ一覧

| パス | ページ | 機能 |
|-----|--------|------|
| `/` | HomePage | ダッシュボード（今日の予定、未完了ToDo、最近のメモ）ウィジェット並び替え対応 |
| `/memos` | MemosPage | メモ一覧、タグフィルタ |
| `/memos/:id` | MemoEditPage | CodeMirror 6エディタ、ライブプレビュー、メモ間リンク |
| `/todo` | TodoPage | インラインタスク作成、詳細パネル、完了フィルタ |
| `/calendar` | CalendarPage | 月表示、予定作成/編集ダイアログ |
| `/tools` | ToolsPage | カテゴリグルーピング、ヘルスチェック |
| `/settings` | SettingsPage | プロフィール/外観/セキュリティ/データの4タブ |
| `/login` | LoginPage | メール+パスワード認証 |

### 状態管理

Zustandで以下をグローバル管理:

- `user` — ログインユーザー情報
- `theme` — テーマ設定
- `sidebarOpen` — サイドバー状態

ページ固有のデータは各ページコンポーネントのローカルstateで管理。

---

## 検索

pg_bigmを使った2-gram全文検索。

- メモ: title + content
- ToDo: title + description
- 予定: title + description

`ILIKE` による部分一致検索をベースに、pg_bigmインデックスで高速化。日本語の分かち書きなしで動作する。

---

## エクスポート/インポート

### エクスポート

| 形式 | エンドポイント | 内容 |
|------|-------------|------|
| JSON | `GET /export/json` | 全データをJSON構造で出力（version, exported_at付き） |
| Markdown ZIP | `GET /export/markdown` | memos/*.md（frontmatter付き）、todos.md、events.md、tools.md |

### インポート

| 形式 | エンドポイント | 内容 |
|------|-------------|------|
| iCal | `POST /import/ical` | RFC 5545形式の.icsファイル。VEVENTをカレンダー予定として取り込み |

iCalパーサーは継続行、日付のみ/日時/タイムゾーン付きの各フォーマット、RRULE（DAILY/WEEKLY/MONTHLY/YEARLY）に対応。

---

## 設計原則

1. **REST APIが唯一の窓口** — Web UIもCLIも外部連携もすべて同じAPIを通る
2. **入力の摩擦ゼロ** — 思いついた瞬間に書き始められる速度
3. **検索は最重要機能** — 「あれどこに書いたっけ」をゼロにする
4. **Markdownは編集とプレビューを明確に分離** — WYSIWYGは採用しない
5. **セルフホストで全機能動作** — 外部サービス依存なし
6. **過剰な機能を入れない** — グループウェア的機能は対象外

---

## ディレクトリ構成

```
konbu/
├── cmd/
│   ├── server/          # APIサーバーエントリポイント
│   └── konbu/           # CLIエントリポイント
├── internal/
│   ├── config/          # 環境変数・設定
│   ├── handler/         # HTTPハンドラ
│   ├── middleware/       # 認証・ログ
│   ├── model/           # リクエスト・レスポンス構造体
│   ├── repository/      # DBアクセス（sqlc）
│   ├── service/         # ビジネスロジック
│   └── client/          # APIクライアント（CLI用）
├── sql/
│   ├── schema.sql       # DDL（参照用）
│   ├── migrations/      # マイグレーションファイル
│   └── queries/         # sqlcクエリ定義
├── web/
│   └── frontend/        # React + Vite SPA
├── docker/
│   └── Dockerfile       # マルチステージビルド
├── docs/                # 設計・API仕様
├── docker-compose.yml   # 開発用
└── docker-compose.prod.yml  # 本番用（Traefik連携）
```
