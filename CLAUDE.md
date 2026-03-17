# CLAUDE.md — konbu 開発ガイド

## プロジェクト概要

konbu はパーソナルワークスペース。メモ・ToDo・カレンダー・ツールランチャーを
REST API + Web UI + CLI で提供する。OSS (MIT) + クラウド版の2形態。

## 技術スタック

- バックエンド: Go 1.25+ / chi v5 / sqlc
- フロントエンド: React 19 + TypeScript / Vite / shadcn/ui / Zustand
- DB: PostgreSQL 16+（pg_trgm で全文検索）
- CLI: cobra
- コンテナ: Docker（マルチステージビルド、scratch ベース）
- i18n: i18next（日本語・英語）

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
│   ├── 00-index.md      # ドキュメントインデックス
│   ├── 01-overview/     # 概要・目的・スコープ
│   ├── 02-architecture/ # 設計・構成・技術スタック
│   ├── 03-details/      # API・データモデル・UI・フロー
│   └── schema.sql       # DDL参照用（全マイグレーション統合版）
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

- Web UI: メール+パスワードでログイン → HMAC署名セッションCookie
- CLI: `Authorization: Bearer <api-key>` → api_keys テーブルの key_hash と照合
- 開発環境: `DEV_USER` 環境変数で自動ログイン
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
| `SESSION_SECRET` | Yes | `konbu-dev-secret-change-me` | セッション署名キー |
| `PORT` | No | `8080` | API サーバーポート |
| `DEV_USER` | No | — | 開発用自動ログインユーザー（メール形式） |
| `OPEN_REGISTRATION` | No | — | `true` で誰でもアカウント作成可能（Cloud版向け） |

## やらないこと

- gRPC（REST のみ）
- GraphQL
- WebSocket
- ORM（sqlc でコード生成。手書き SQL を基本とする）

## セッションメモ（2026-03-15〜16）

### 完了した作業（3/15-16）
- #3 AIチャット（BYOK、OpenAI/Anthropic、SSEストリーミング、15+2ツール）
- #5 MCP Server（`konbu mcp` サブコマンド、stdio、18ツール）
- #6 画像添付（Cloudflare R2、ドラッグ&ドロップ/ペースト）
- #8 レート制限（Web UI免除、CLI/API 100req/分）
- #12 ツール機能改善（favicon自動取得、定期リフレッシュ）
- #13 旧Supabase削除
- #14 Google OAuth設定（メール正規化修正含む）
- #15 GitHub Sponsors設定
- #16 セルフホスト版停止
- #17 パフォーマンス改善（Monaco lazy load、Cache-Control）
- #18 ツールのドラッグ&ドロップ並び替え
- #19 課金導線（Ko-fi連携、Webhook、アプリ内Ko-fi埋め込み）
- #20 AIチャット改善（web_search/web_fetch、Shift+Enter、SSEパース修正）
- #21 スポンサー向けUI改善
- CD: GitHub Actions → Fly.io自動デプロイ
- 自動マイグレーション（サーバー起動時にSQL適用）
- ランディングページ（konbu.codenica.dev、Cloudflare Pages）
- ダークモード切り替え（ライト/ダーク/システム）
- PWA対応（manifest.json）
- iCal URL公開（GET /calendar.ics?token=xxx）
- セキュリティ修正（XSS: DOMPurify、セキュリティヘッダー、npm audit）
- Admin全機能アクセス（is_admin=true → Sponsor機能利用可）
- ヘルスチェック機能削除（不要・誤検知多い）
- Supabase Pooler接続に切り替え（IPv6直接接続廃止対応）
- SearXNGセルフホスト（apps-vps、チャットのWeb検索用）
- ディスク容量対策（Docker→/home移動）
- hotateコピペ修正（tmux mouse mode対応）
- ブックマーク移植（CLI経由、20件カテゴリ付き）
- コード剽窃・ライセンスチェック（問題なし）

### 判断・決定事項
- 課金: Ko-fi（寄付型、特商法回避）→ 規模拡大時にStripe移行
- ランディング: 非エンジニア向け、OSS/GitHub/セルフホストは訴求しない
- Web検索: 公開SearXNG → セルフホスト（apps-vps）に切り替え
- ヘルスチェック: 削除（Uptime Kumaに任せる）
- レート制限: Web UI(Cookie)は免除、CLI/API(Bearer)のみ制限

### 未解決の問題
- Ko-fi Webhookの実際の動作確認（テスト支払い未実施）
- チャットのメッセージ消失が完全に解消されたか未確認
- デバッグログ（favicon refresh）がまだ残っている

### 次回の優先事項
1. X/Bluesky アカウント作成・宣伝開始
2. #7 リマインダー通知（Brevoメール + ブラウザ内バナー）
3. #22 スマホUI残り（MemoEditPageツールバー等）
4. #9 CIにテスト追加
5. Ko-fi Webhook動作確認
