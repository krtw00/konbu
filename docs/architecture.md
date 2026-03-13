# Architecture

## 全体像

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

## レイヤー構成

```
handler     HTTP入出力。バリデーション、レスポンス整形
service     ビジネスロジック。トランザクション管理、タグupsert
repository  DBアクセス。sqlc生成コード + カスタムクエリ
```

handler → service → repository の一方向依存。逆方向の参照はしない。

## 技術スタック

| レイヤー | 技術 |
|---------|------|
| バックエンド | Go / chi / sqlc / cobra |
| フロントエンド | React 19 / TypeScript / shadcn/ui / Zustand / Vite |
| エディタ | CodeMirror 6 (@uiw/react-codemirror) |
| i18n | i18next / react-i18next (日本語・英語) |
| DB | PostgreSQL 16+ |
| 全文検索 | pg_bigm (2-gram、日本語対応) |
| コンテナ | Docker (マルチステージビルド、scratchベース) |

## モジュール一覧

| モジュール | handler | service | 機能 |
|-----------|---------|---------|------|
| auth | auth_handler | auth_service | ユーザー登録/ログイン/設定 |
| memo | memo_handler | memo_service | メモCRUD |
| todo | todo_handler | todo_service | ToDo CRUD、完了/未完了 |
| event | event_handler | event_service | カレンダー予定CRUD |
| tool | tool_handler | tool_service | ツールCRUD、favicon、ヘルスチェック |
| tag | tag_handler | tag_service | タグCRUD、暗黙的upsert |
| search | search_handler | search_service | 横断全文検索 |
| export | export_handler | export_service | JSON/Markdown ZIPエクスポート |
| import | import_handler | import_service | iCalインポート |

## フロントエンド

```
web/frontend/src/
  pages/        ページコンポーネント（ルートごと）
  components/   共通UI (shadcn/ui)
  stores/       状態管理 (Zustand)
  lib/          APIクライアント、ユーティリティ
  i18n/         翻訳ファイル (en.json, ja.json)
```

Zustandでグローバル管理するのは `user`, `theme`, `sidebarOpen` のみ。ページ固有のデータはローカルstate。

## ディレクトリ構成

```
konbu/
├── cmd/
│   ├── server/        # APIサーバー
│   └── konbu/         # CLI
├── internal/
│   ├── config/        # 環境変数
│   ├── handler/       # HTTPハンドラ
│   ├── middleware/     # 認証・ログ
│   ├── model/         # 構造体
│   ├── repository/    # DB (sqlc)
│   ├── service/       # ビジネスロジック
│   └── client/        # APIクライアント (CLI用)
├── sql/
│   ├── migrations/    # マイグレーション
│   └── queries/       # sqlcクエリ
├── web/frontend/      # React SPA
├── docker/            # Dockerfile
└── docs/              # ドキュメント
```

## 設計原則

1. **REST APIが唯一の窓口** — UI・CLI・外部連携すべて同じAPI
2. **入力の摩擦ゼロ** — 思いついた瞬間に書き始められる
3. **検索は最重要機能** — 横断全文検索で「どこに書いたか」を解決
4. **Markdownは編集とプレビューを分離** — WYSIWYGは使わない
5. **セルフホストで全機能** — 外部サービス依存なし
6. **過剰に作らない** — 必要なものだけ入れる
