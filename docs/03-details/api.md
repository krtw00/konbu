---
depends_on:
  - ../02-architecture/structure.md
  - ./data-model.md
tags: [details, api, endpoints, rest]
ai_summary: "konbuのREST APIエンドポイント一覧・認証・レスポンス形式・共通仕様を定義"
---

# API設計

> **Status**: Active | 最終更新: 2026-03-14

本ドキュメントは、konbuのREST API設計を定義する。

---

## API概要

| 項目 | 内容 |
|------|------|
| ベースURL | `/api/v1` |
| 認証方式 | Session Cookie (Web UI) / Bearer Token (CLI・外部連携) |
| レスポンス形式 | JSON |

---

## エンドポイント一覧

### Auth（公開）

| メソッド | パス | 説明 |
|----------|------|------|
| POST | `/auth/register` | ユーザー登録 |
| POST | `/auth/login` | ログイン（Session Cookie発行） |
| POST | `/auth/logout` | ログアウト（Cookie削除） |
| GET | `/auth/setup-status` | 初期セットアップ状態確認 |

### Auth（認証済み）

| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/auth/me` | 現在のユーザー情報取得 |
| PUT | `/auth/me` | プロフィール更新 |
| POST | `/auth/change-password` | パスワード変更 |
| GET | `/auth/settings` | ユーザー設定取得 (JSONB) |
| PUT | `/auth/settings` | ユーザー設定更新 |

### API Keys

| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/api-keys` | APIキー一覧 |
| POST | `/api-keys` | APIキー発行（生キーは1度だけ返却） |
| DELETE | `/api-keys/:id` | APIキー削除 |

### Memos

| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/memos` | メモ一覧 |
| POST | `/memos` | メモ作成 |
| GET | `/memos/:id` | メモ詳細 |
| PUT | `/memos/:id` | メモ更新 |
| DELETE | `/memos/:id` | メモ削除（論理削除） |

追加フィルタ: `type` (`markdown` / `table`)

### ToDos

| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/todos` | ToDo一覧 |
| POST | `/todos` | ToDo作成 |
| GET | `/todos/:id` | ToDo詳細 |
| PUT | `/todos/:id` | ToDo更新 |
| PATCH | `/todos/:id/done` | 完了にする |
| PATCH | `/todos/:id/reopen` | 未完了に戻す |
| DELETE | `/todos/:id` | 削除（論理削除） |

追加フィルタ: `status` (`open` / `done`)

### Calendar Events

| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/events` | イベント一覧 |
| POST | `/events` | イベント作成 |
| GET | `/events/:id` | イベント詳細 |
| PUT | `/events/:id` | イベント更新 |
| DELETE | `/events/:id` | 削除（論理削除） |

追加フィルタ: `from`, `to` (datetime), `month` (`2026-03`形式)

### Tools

| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/tools` | ツール一覧（sort_order順） |
| POST | `/tools` | ツール追加（favicon自動取得） |
| PUT | `/tools/:id` | ツール更新 |
| DELETE | `/tools/:id` | ツール削除（論理削除） |
| PUT | `/tools/reorder` | 並び替え |
| POST | `/tools/refresh-icons` | 全ツールのfavicon再取得 |
| POST | `/tools/health-check` | 全ツールのURL疎通確認 |

### Tags

| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/tags` | タグ一覧 |
| POST | `/tags` | タグ作成 |
| PUT | `/tags/:id` | タグ名変更 |
| DELETE | `/tags/:id` | タグ削除（論理削除） |

### Search

| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/search?q=...` | メモ・ToDo・予定の横断全文検索 |

### Export

| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/export/json` | 全データをJSONでダウンロード |
| GET | `/export/markdown` | 全データをMarkdown ZIPでダウンロード |

### Import

| メソッド | パス | 説明 |
|----------|------|------|
| POST | `/import/ical` | iCalファイル(.ics)をインポート（最大10MB） |

---

## 共通仕様

### 認証

| 方式 | ヘッダー / Cookie | 用途 |
|------|------------------|------|
| APIキー | `Authorization: Bearer <api-key>` | CLI / 外部連携 |
| セッションCookie | `konbu_session` (HMAC-SHA256署名) | Web UI |
| 開発モード | `DEV_USER` 環境変数 | ローカル開発 |

### エラーレスポンス

| フィールド | 型 | 説明 |
|------------|-----|------|
| error.code | string | エラーコード (`not_found`, `validation_error` 等) |
| error.message | string | エラーメッセージ |

```json
{
  "error": {
    "code": "not_found",
    "message": "memo not found"
  }
}
```

### レスポンス形式

一覧:
```json
{
  "data": [...],
  "total": 100,
  "limit": 20,
  "offset": 0
}
```

単体:
```json
{
  "data": { ... }
}
```

### 共通クエリパラメータ（一覧系）

| パラメータ | 型 | デフォルト | 説明 |
|---|---|---|---|
| `limit` | int | 20 | 取得件数（最大100） |
| `offset` | int | 0 | オフセット |
| `sort` | string | `created_at:desc` | ソート (`field:asc` or `field:desc`) |
| `q` | string | -- | テキスト検索 (ILIKE + pg_bigm) |
| `tag` | string | -- | タグ名で絞り込み |

---

## 関連ドキュメント

- [data-model.md](./data-model.md) - データモデル
- [flows.md](./flows.md) - 主要フロー
