# konbu REST API 仕様

> 更新: 2026-03-13

## 共通仕様

### ベースパス

```
/api/v1
```

### 認証

| 方式 | ヘッダー / Cookie | 用途 |
|------|------------------|------|
| APIキー | `Authorization: Bearer <api-key>` | CLI / 外部連携 |
| セッションCookie | `konbu_session` (HMAC署名) | Web UI |
| 開発モード | `DEV_USER` 環境変数 | ローカル開発 |

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

エラー:
```json
{
  "error": {
    "code": "not_found",
    "message": "memo not found"
  }
}
```

### 共通クエリパラメータ（一覧系）

| パラメータ | 型 | デフォルト | 説明 |
|---|---|---|---|
| `limit` | int | 20 | 取得件数（最大100） |
| `offset` | int | 0 | オフセット |
| `sort` | string | `created_at:desc` | ソート（`field:asc` or `field:desc`） |
| `q` | string | — | テキスト検索（ILIKE + pg_bigm） |
| `tag` | string | — | タグ名で絞り込み |

---

## Auth

### 公開エンドポイント（認証不要）

| メソッド | パス | 説明 |
|---|---|---|
| POST | `/auth/register` | ユーザー登録（メール+パスワード） |
| POST | `/auth/login` | ログイン（セッションCookie発行） |
| POST | `/auth/logout` | ログアウト（Cookie削除） |
| GET | `/auth/setup-status` | 初期セットアップ状態の確認 |

### 認証済みエンドポイント

| メソッド | パス | 説明 |
|---|---|---|
| GET | `/auth/me` | 現在のユーザー情報 |
| PUT | `/auth/me` | プロフィール更新（name） |
| POST | `/auth/change-password` | パスワード変更 |
| GET | `/auth/settings` | ユーザー設定取得（JSONB） |
| PUT | `/auth/settings` | ユーザー設定更新 |

### POST /auth/register

```json
{
  "email": "user@example.com",
  "password": "password",
  "name": "User Name"
}
```

### POST /auth/login

```json
{
  "email": "user@example.com",
  "password": "password"
}
```

### POST /auth/change-password

```json
{
  "old_password": "current",
  "new_password": "new"
}
```

---

## API Keys

| メソッド | パス | 説明 |
|---|---|---|
| GET | `/api-keys` | APIキー一覧 |
| POST | `/api-keys` | APIキー発行（生キーは1度だけ返却） |
| DELETE | `/api-keys/:id` | APIキー削除 |

### POST /api-keys

リクエスト:
```json
{
  "name": "cli-home"
}
```

レスポンス:
```json
{
  "data": {
    "id": "uuid",
    "name": "cli-home",
    "key": "konbu_xxxxxxxxxxxxxxxxxxxx",
    "created_at": "2026-03-13T00:00:00Z"
  }
}
```

---

## Tags

| メソッド | パス | 説明 |
|---|---|---|
| GET | `/tags` | タグ一覧 |
| POST | `/tags` | タグ作成 |
| PUT | `/tags/:id` | タグ名変更 |
| DELETE | `/tags/:id` | タグ削除（論理削除） |

---

## Memos

| メソッド | パス | 説明 |
|---|---|---|
| GET | `/memos` | メモ一覧 |
| POST | `/memos` | メモ作成 |
| GET | `/memos/:id` | メモ詳細 |
| PUT | `/memos/:id` | メモ更新 |
| DELETE | `/memos/:id` | メモ削除（論理削除） |

### 追加フィルタ

| パラメータ | 型 | 説明 |
|---|---|---|
| `type` | string | `markdown` or `table` |

### POST /memos

```json
{
  "title": "メモタイトル",
  "type": "markdown",
  "content": "# 見出し\n\n本文...",
  "tags": ["tag1", "tag2"]
}
```

### GET /memos/:id レスポンス

```json
{
  "data": {
    "id": "uuid",
    "title": "メモタイトル",
    "type": "markdown",
    "content": "# 見出し\n\n本文...",
    "tags": [
      {"id": "uuid", "name": "tag1"}
    ],
    "created_at": "2026-03-13T00:00:00Z",
    "updated_at": "2026-03-13T00:00:00Z"
  }
}
```

---

## ToDos

| メソッド | パス | 説明 |
|---|---|---|
| GET | `/todos` | ToDo一覧 |
| POST | `/todos` | ToDo作成 |
| GET | `/todos/:id` | ToDo詳細 |
| PUT | `/todos/:id` | ToDo更新 |
| PATCH | `/todos/:id/done` | 完了にする |
| PATCH | `/todos/:id/reopen` | 未完了に戻す |
| DELETE | `/todos/:id` | 削除（論理削除） |

### 追加フィルタ

| パラメータ | 型 | 説明 |
|---|---|---|
| `status` | string | `open` or `done` |

### POST /todos

```json
{
  "title": "タスク名",
  "description": "詳細メモ",
  "status": "open",
  "due_date": "2026-03-15",
  "tags": ["仕事"]
}
```

---

## Calendar Events

| メソッド | パス | 説明 |
|---|---|---|
| GET | `/events` | イベント一覧 |
| POST | `/events` | イベント作成 |
| GET | `/events/:id` | イベント詳細 |
| PUT | `/events/:id` | イベント更新 |
| DELETE | `/events/:id` | 削除（論理削除） |

### 追加フィルタ

| パラメータ | 型 | 説明 |
|---|---|---|
| `from` | datetime | 開始日時以降 |
| `to` | datetime | 開始日時以前 |
| `month` | string | `2026-03` 形式で月指定 |

### POST /events

```json
{
  "title": "予定名",
  "start_at": "2026-03-13T09:00:00+09:00",
  "end_at": "2026-03-13T17:00:00+09:00",
  "all_day": false,
  "description": "詳細",
  "tags": ["仕事"]
}
```

---

## Tools

| メソッド | パス | 説明 |
|---|---|---|
| GET | `/tools` | ツール一覧（sort_order順） |
| POST | `/tools` | ツール追加（favicon自動取得） |
| PUT | `/tools/:id` | ツール更新 |
| DELETE | `/tools/:id` | ツール削除（論理削除） |
| PUT | `/tools/reorder` | 並び替え |
| POST | `/tools/refresh-icons` | 全ツールのfavicon再取得 |
| POST | `/tools/health-check` | 全ツールのURL疎通確認 |

### POST /tools

```json
{
  "name": "GitHub",
  "url": "https://github.com",
  "icon": "",
  "category": "Dev"
}
```

---

## Search（横断検索）

| メソッド | パス | 説明 |
|---|---|---|
| GET | `/search` | メモ・ToDo・予定を横断検索 |

### パラメータ

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `q` | string | Yes | 検索クエリ |

### レスポンス

```json
{
  "data": [
    {
      "type": "memo",
      "id": "uuid",
      "title": "該当メモ",
      "snippet": "...一致箇所..."
    },
    {
      "type": "todo",
      "id": "uuid",
      "title": "該当タスク",
      "snippet": "..."
    }
  ]
}
```

---

## Export

| メソッド | パス | 説明 |
|---|---|---|
| GET | `/export/json` | 全データをJSONでダウンロード |
| GET | `/export/markdown` | 全データをMarkdown ZIPでダウンロード |

### JSON形式

```json
{
  "version": 1,
  "exported_at": "2026-03-13T...",
  "memos": [...],
  "todos": [...],
  "events": [...],
  "tools": [...]
}
```

### Markdown ZIP構造

```
memos/
  メモタイトル.md       # frontmatter (tags, created_at) + content
  ...
todos.md               # チェックリスト形式
events.md              # テーブル形式
tools.md               # テーブル形式
```

---

## Import

| メソッド | パス | 説明 |
|---|---|---|
| POST | `/import/ical` | iCalファイル(.ics)をインポート |

マルチパートフォームでファイルをアップロード（フィールド名: `file`、最大10MB）。

RFC 5545のVEVENTを解析し、カレンダー予定として取り込む。DTSTART/DTEND（日付のみ/日時/タイムゾーン付き）、RRULE（DAILY/WEEKLY/MONTHLY/YEARLY）に対応。
