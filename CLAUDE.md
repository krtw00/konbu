# CLAUDE.md — konbu 開発ガイド

## プロジェクト概要

konbu はセルフホスト型の個人ツール基盤。メモ・ToDo・カレンダー・ツールランチャーを
REST API + Web UI + CLI で提供する。

## 技術スタック

- 言語: Go 1.22+
- DB: PostgreSQL 16+（pg_bigm で日本語全文検索）
- ルーター: echo or chi（後で確定）
- SQL: sqlc でコード生成
- CLI: cobra
- コンテナ: Docker（マルチステージビルド、scratch ベース）

## ディレクトリ構成

```
konbu/
├── CLAUDE.md
├── README.md
├── go.mod
├── go.sum
├── cmd/
│   ├── server/          # API サーバーのエントリポイント
│   │   └── main.go
│   └── konbu/           # CLI のエントリポイント
│       └── main.go
├── internal/
│   ├── config/          # 環境変数・設定読み込み
│   ├── handler/         # HTTP ハンドラ（エンドポイントごと）
│   ├── middleware/       # 認証・ログ・エラーハンドリング
│   ├── model/           # 構造体定義（リクエスト・レスポンス）
│   ├── repository/      # sqlc 生成コード + カスタムクエリ
│   ├── service/         # ビジネスロジック
│   └── testutil/        # テスト用ヘルパー
├── sql/
│   ├── schema.sql       # DDL（docs/schema.sql のコピーではなく本体）
│   ├── migrations/      # マイグレーションファイル
│   └── queries/         # sqlc 用クエリファイル
│       ├── users.sql
│       ├── memos.sql
│       ├── todos.sql
│       ├── events.sql
│       ├── tags.sql
│       ├── tools.sql
│       └── search.sql
├── docs/
│   ├── design.md        # 構想メモ
│   ├── api.md           # REST API 仕様
│   └── schema.sql       # DDL（参照用）
├── docker/
│   └── Dockerfile
└── docker-compose.yml
```

## 命名規則

### Go コード

- パッケージ名: 小文字単一語（`handler`, `service`, `repository`）
- ファイル名: スネークケース（`memo_handler.go`, `calendar_event.go`）
- 構造体: パスカルケース（`Memo`, `CreateMemoRequest`）
- 関数: パスカルケース（公開）、キャメルケース（非公開）
- 変数: キャメルケース
- 定数: パスカルケース（`StatusOpen`, `TypeMarkdown`）
- テストファイル: `*_test.go`

### SQL

- テーブル名: スネークケース複数形（`memos`, `memo_tags`, `calendar_events`）
- カラム名: スネークケース（`user_id`, `created_at`, `deleted_at`）
- インデックス名: `idx_{table}_{columns}`
- マイグレーション: `NNNN_description.up.sql` / `NNNN_description.down.sql`

### API

- パス: ケバブケース複数形（`/api/v1/memos`, `/api/v1/api-keys`）
- JSON フィールド: スネークケース（`created_at`, `user_id`, `table_columns`）

## コーディングルール

### アーキテクチャ

- レイヤー: handler → service → repository の 3 層
- handler: HTTP の入出力のみ。バリデーション、レスポンス整形
- service: ビジネスロジック。トランザクション管理
- repository: DB アクセスのみ。sqlc 生成コードを基本とし、複雑なクエリのみカスタム
- 各レイヤーはインターフェースで依存（テスト容易性のため）

### データフロー

```
Request → middleware(認証) → handler → service → repository → PostgreSQL
                              ↓
                           Response (JSON)
```

### エラーハンドリング

- アプリケーションエラーは独自の error 型で定義（`internal/apperror/`）
- handler 層で HTTP ステータスコードにマッピング
- エラーレスポンスは統一形式: `{"error": {"code": "xxx", "message": "xxx"}}`

### 論理削除

- 削除 API は `deleted_at = now()` をセット
- 全 SELECT クエリに `WHERE deleted_at IS NULL` を付与
- 物理削除は `/trash` エンドポイント経由のみ

### 認証

- Web UI: `X-Forwarded-User` ヘッダーからメールアドレス取得 → users テーブル参照（なければ自動登録）
- CLI/bot: `Authorization: Bearer <key>` → api_keys テーブルの key_hash と照合
- 全エンドポイントでユーザーコンテキストを注入、自分のデータのみアクセス可能

### タグ

- メモ・ToDo・イベント作成時に `tags: ["name1", "name2"]` を受け取る
- 存在しないタグ名は自動作成（暗黙的 upsert）
- 中間テーブルの付け替えは service 層で処理

### テスト

- テストは `*_test.go` に書く
- DB を使うテストは testutil でテスト用 DB をセットアップ
- handler のテストは httptest + モック service
- service のテストは モック repository
- repository のテストは実 DB（テスト用 PostgreSQL コンテナ）

## コマンド

```bash
# 開発
go run ./cmd/server          # API サーバー起動
go run ./cmd/konbu           # CLI 実行

# コード生成
sqlc generate                # SQL → Go コード生成

# テスト
go test ./...                # 全テスト実行
go test ./internal/handler/  # handler のみ

# ビルド
go build -o bin/server ./cmd/server
go build -o bin/konbu ./cmd/konbu

# Docker
docker compose up -d         # 全サービス起動
docker compose up -d postgres # DB のみ起動

# マイグレーション
go run ./cmd/server migrate up
go run ./cmd/server migrate down
```

## 環境変数

| 変数 | 必須 | デフォルト | 説明 |
|---|---|---|---|
| `DATABASE_URL` | Yes | — | PostgreSQL 接続文字列 |
| `PORT` | No | `8080` | API サーバーポート |
| `ADMIN_EMAIL` | No | — | 管理者メールアドレス（指定時はこのユーザーを管理者にする） |
| `ALLOWED_EMAILS` | No | `*` | 許可するメールアドレス（カンマ区切り、`*` で全許可） |

## やらないこと

- gRPC（REST のみ）
- GraphQL
- WebSocket（ポータルには不要。WebSSH は別サービス）
- ORM（sqlc でコード生成。手書き SQL を基本とする）
- フロントエンドのフレームワーク選定（別途決定）
