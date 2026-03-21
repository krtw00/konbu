---
depends_on:
  - ../02-architecture/structure.md
  - ./data-model.md
tags: [details, api, endpoints, rest]
ai_summary: "konbuのREST APIエンドポイント一覧・認証・レスポンス形式・共通仕様を定義"
---

# API設計

> **Status**: Active | 最終更新: 2026-03-18

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
| GET | `/auth/providers` | 有効なログイン手段一覧 |
| GET | `/auth/google/login` | Google OAuth 開始 |
| GET | `/auth/google/callback` | Google OAuth コールバック |

### Public（公開）

| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/public/:token` | 共有リンクの閲覧用データ取得（read-only） |
| GET | `/published/:resourceType/:slug` | publish metadata の取得（slug lookup） |

### Auth（認証済み）

| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/auth/me` | 現在のユーザー情報取得 |
| PUT | `/auth/me` | プロフィール更新 |
| POST | `/auth/change-password` | パスワード変更 |
| GET | `/auth/settings` | ユーザー設定取得 (JSONB) |
| PUT | `/auth/settings` | ユーザー設定更新 |
| POST | `/auth/delete-account` | アカウント削除 |

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
| GET | `/memos/:id/rows` | テーブル行一覧 |
| POST | `/memos/:id/rows` | テーブル行追加 |
| POST | `/memos/:id/rows/batch` | テーブル行一括追加 |
| GET | `/memos/:id/rows/export` | CSVエクスポート |
| PUT | `/memos/:id/rows/:rowId` | テーブル行更新 |
| DELETE | `/memos/:id/rows/:rowId` | テーブル行削除 |

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

### Calendars

| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/calendars` | カレンダー一覧 |
| POST | `/calendars` | カレンダー作成 |
| GET | `/calendars/:id` | カレンダー詳細（メンバー含む） |
| PUT | `/calendars/:id` | カレンダー更新 |
| DELETE | `/calendars/:id` | カレンダー削除 |
| POST | `/calendars/join/:token` | 共有リンクで参加 |
| POST | `/calendars/:id/share-link` | 共有リンク生成 |
| DELETE | `/calendars/:id/share-link` | 共有リンク無効化 |
| POST | `/calendars/:id/members` | メンバー追加 |
| PUT | `/calendars/:id/members/:uid` | メンバー権限/色更新 |
| DELETE | `/calendars/:id/members/:uid` | メンバー削除 |
| GET | `/calendar.ics` | iCalフィード取得（token認証） |

### Public Shares

| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/public-shares/:resourceType/:id` | リソースの共有リンク取得 |
| POST | `/public-shares/:resourceType/:id` | 閲覧専用の共有リンク作成 |
| DELETE | `/public-shares/:resourceType/:id` | 共有リンク削除 |

対応する `resourceType`: `memo`, `todo`, `calendar`, `event`

補足: `tool` の public share は旧実装として backend に残る可能性があるが、現行仕様では対象外とする。

### Publishes

| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/publishes/:resourceType/:id` | リソースの publish metadata 取得 |
| PUT | `/publishes/:resourceType/:id` | slug / title / description / visibility を upsert |
| DELETE | `/publishes/:resourceType/:id` | publish metadata 削除 |

対応する `resourceType`: `memo`, `event`, `calendar`

補足:

- 現段階の publish API は metadata 管理と slug lookup まで
- 実際の公開ページ本文や表示 UI は後続 issue (`memo publish`, `event / calendar publish`) で扱う

### Tools

| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/tools` | ツール一覧（sort_order順） |
| POST | `/tools` | ツール追加（favicon自動取得） |
| PUT | `/tools/:id` | ツール更新 |
| DELETE | `/tools/:id` | ツール削除（論理削除） |
| PUT | `/tools/reorder` | 並び替え |
| POST | `/tools/refresh-icons` | 全ツールのfavicon再取得（手動） |
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
| GET | `/search?q=...` | メモ・ToDo・予定・ツールの横断検索 |

追加パラメータ: `type`, `tag`, `from`, `to`, `limit`, `offset`

### Chat

| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/chat/sessions` | チャットセッション一覧 |
| POST | `/chat/sessions` | チャットセッション作成 |
| GET | `/chat/sessions/:id` | チャットセッション詳細 |
| PUT | `/chat/sessions/:id` | セッション名更新 |
| DELETE | `/chat/sessions/:id` | セッション削除 |
| POST | `/chat/sessions/:id/messages` | メッセージ送信（SSEストリーム） |
| GET | `/chat/config` | AI設定取得 |
| PUT | `/chat/config` | AI設定保存 |

### Attachments

| メソッド | パス | 説明 |
|----------|------|------|
| POST | `/attachments` | 画像添付アップロード（Sponsor/Admin） |
| GET | `/attachments/*` | 添付ファイル取得 |

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
| セッションCookie | `__session` (HMAC-SHA256署名) | Web UI |
| 開発モード | `DEV_USER` 環境変数 | ローカル開発 |
| iCal feed | `GET /api/v1/calendar.ics?token=...` | 外部カレンダー購読 |
| 共有リンク | `GET /api/v1/public/:token` | ログイン不要の閲覧専用ページ |
| publish metadata | `GET /api/v1/published/:resourceType/:slug` | ログイン不要の slug lookup |

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
| `q` | string | -- | テキスト検索 (ILIKE + pg_trgm) |
| `tag` | string | -- | タグ名で絞り込み |

---

## 関連ドキュメント

- [data-model.md](./data-model.md) - データモデル
- [flows.md](./flows.md) - 主要フロー
