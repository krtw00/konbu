# konbu REST API 仕様

## 共通仕様

### ベースパス

```
/api/v1
```

### 認証

| アクセス元 | 方式 | ヘッダー |
|---|---|---|
| Web UI | ForwardAuth Cookie（Traefik が処理） | `X-Forwarded-User` からユーザー識別 |
| CLI / bot | API キー | `Authorization: Bearer <api-key>` |

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
| `q` | string | — | テキスト検索（pg_bigm） |
| `tag` | string | — | タグ名で絞り込み（複数指定はカンマ区切り） |

---

## Auth

| メソッド | パス | 説明 |
|---|---|---|
| GET | `/auth/me` | 現在のユーザー情報を取得 |
| PUT | `/auth/me` | プロフィール更新（name） |

---

## API Keys

| メソッド | パス | 説明 |
|---|---|---|
| GET | `/api-keys` | 自分の API キー一覧 |
| POST | `/api-keys` | API キー発行（レスポンスで生キーを1度だけ返す） |
| DELETE | `/api-keys/:id` | API キー無効化（物理削除） |

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
    "created_at": "2026-03-12T00:00:00Z"
  }
}
```

---

## Tags

| メソッド | パス | 説明 |
|---|---|---|
| GET | `/tags` | タグ一覧（使用数付き） |
| POST | `/tags` | タグ作成 |
| PUT | `/tags/:id` | タグ名変更 |
| DELETE | `/tags/:id` | タグ削除（論理削除） |

### GET /tags レスポンス

```json
{
  "data": [
    {
      "id": "uuid",
      "name": "flask",
      "counts": {
        "memos": 5,
        "todos": 2,
        "events": 0
      }
    }
  ]
}
```

---

## Memos

| メソッド | パス | 説明 |
|---|---|---|
| GET | `/memos` | メモ一覧 |
| POST | `/memos` | メモ作成 |
| GET | `/memos/:id` | メモ詳細 |
| PUT | `/memos/:id` | メモ更新 |
| DELETE | `/memos/:id` | メモ削除（論理削除 → ゴミ箱へ） |

### 追加フィルタ

| パラメータ | 型 | 説明 |
|---|---|---|
| `type` | string | `markdown` or `table` |

### POST /memos（markdown）

リクエスト:
```json
{
  "title": "Flask P5テスト結果",
  "type": "markdown",
  "content": "# 結果\n\n合格。詳細は...",
  "tags": ["flask", "教育"]
}
```

### POST /memos（table）

リクエスト:
```json
{
  "title": "血圧",
  "type": "table",
  "table_columns": [
    {"name": "日付", "type": "date"},
    {"name": "上", "type": "number"},
    {"name": "下", "type": "number"},
    {"name": "脈拍", "type": "number"},
    {"name": "メモ", "type": "text"}
  ],
  "tags": ["log/血圧"]
}
```

### GET /memos/:id レスポンス（markdown）

```json
{
  "data": {
    "id": "uuid",
    "title": "Flask P5テスト結果",
    "type": "markdown",
    "content": "# 結果\n\n合格。詳細は...",
    "tags": [
      {"id": "uuid", "name": "flask"},
      {"id": "uuid", "name": "教育"}
    ],
    "created_at": "2026-03-12T00:00:00Z",
    "updated_at": "2026-03-12T00:00:00Z"
  }
}
```

### GET /memos/:id レスポンス（table）

```json
{
  "data": {
    "id": "uuid",
    "title": "血圧",
    "type": "table",
    "table_columns": [
      {"name": "日付", "type": "date"},
      {"name": "上", "type": "number"},
      {"name": "下", "type": "number"},
      {"name": "脈拍", "type": "number"},
      {"name": "メモ", "type": "text"}
    ],
    "tags": [
      {"id": "uuid", "name": "log/血圧"}
    ],
    "rows": {
      "data": [...],
      "total": 365
    },
    "created_at": "2026-03-12T00:00:00Z",
    "updated_at": "2026-03-12T00:00:00Z"
  }
}
```

---

## Memo Rows

テーブル型メモの行データ操作。

| メソッド | パス | 説明 |
|---|---|---|
| GET | `/memos/:memo_id/rows` | 行一覧（ページネーション対応） |
| POST | `/memos/:memo_id/rows` | 行追加 |
| PUT | `/memos/:memo_id/rows/:id` | 行更新 |
| DELETE | `/memos/:memo_id/rows/:id` | 行削除（論理削除） |
| PUT | `/memos/:memo_id/rows/reorder` | 行並び替え |

### 追加フィルタ

| パラメータ | 型 | 説明 |
|---|---|---|
| `sort` | string | `row_data.日付:desc` 等、JSONB キーでソート |

### POST /memos/:memo_id/rows

リクエスト:
```json
{
  "row_data": {
    "日付": "2026-03-12",
    "上": 128,
    "下": 82,
    "脈拍": 72,
    "メモ": ""
  }
}
```

### PUT /memos/:memo_id/rows/reorder

リクエスト:
```json
{
  "order": ["uuid-row-3", "uuid-row-1", "uuid-row-2"]
}
```

---

## ToDos

| メソッド | パス | 説明 |
|---|---|---|
| GET | `/todos` | ToDo 一覧 |
| POST | `/todos` | ToDo 作成 |
| GET | `/todos/:id` | ToDo 詳細 |
| PUT | `/todos/:id` | ToDo 更新 |
| PATCH | `/todos/:id/done` | 完了にする |
| PATCH | `/todos/:id/reopen` | 未完了に戻す |
| DELETE | `/todos/:id` | 削除（論理削除 → ゴミ箱へ） |

### 追加フィルタ

| パラメータ | 型 | 説明 |
|---|---|---|
| `status` | string | `open` or `done` |
| `due_before` | date | 期限がこの日以前 |
| `due_after` | date | 期限がこの日以降 |
| `overdue` | bool | `true` で期限切れのみ |

### POST /todos

リクエスト:
```json
{
  "title": "DD_00レビュー",
  "description": "",
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
| DELETE | `/events/:id` | 削除（論理削除 → ゴミ箱へ） |

### 追加フィルタ

| パラメータ | 型 | 説明 |
|---|---|---|
| `from` | datetime | 開始日時以降 |
| `to` | datetime | 開始日時以前 |
| `week` | date | 指定日を含む週（月〜日） |
| `month` | string | `2026-03` 形式で月指定 |

### POST /events

リクエスト:
```json
{
  "title": "教育D1開始",
  "start_at": "2026-03-12T09:00:00+09:00",
  "end_at": "2026-03-12T17:00:00+09:00",
  "all_day": false,
  "tags": ["教育"]
}
```

---

## Tools

| メソッド | パス | 説明 |
|---|---|---|
| GET | `/tools` | ツール一覧（sort_order 順） |
| POST | `/tools` | ツール追加 |
| PUT | `/tools/:id` | ツール更新 |
| DELETE | `/tools/:id` | ツール削除（論理削除） |
| PUT | `/tools/reorder` | 並び替え |

### POST /tools

リクエスト:
```json
{
  "name": "WebSSH",
  "url": "https://webssh.example.com",
  "icon": "terminal"
}
```

---

## Search（横断検索）

| メソッド | パス | 説明 |
|---|---|---|
| GET | `/search` | メモ・ToDo・カレンダーを横断検索 |

### パラメータ

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `q` | string | Yes | 検索クエリ |
| `scope` | string | No | `memos`,`todos`,`events`（カンマ区切り、省略時は全対象） |
| `limit` | int | No | 各エンティティの最大件数（デフォルト10） |

### レスポンス

```json
{
  "data": {
    "memos": {
      "items": [...],
      "total": 12
    },
    "todos": {
      "items": [...],
      "total": 3
    },
    "events": {
      "items": [...],
      "total": 1
    }
  },
  "query": "Flask"
}
```

---

## Trash（ゴミ箱）

| メソッド | パス | 説明 |
|---|---|---|
| GET | `/trash` | ゴミ箱内の全アイテム一覧 |
| POST | `/trash/:type/:id/restore` | アイテムを復元 |
| DELETE | `/trash/:type/:id` | アイテムを物理削除 |
| DELETE | `/trash` | ゴミ箱を空にする（全件物理削除） |

### パラメータ

| パラメータ | 型 | 説明 |
|---|---|---|
| `type` | string | `memos`, `todos`, `events`, `tools`, `tags` |

### GET /trash レスポンス

```json
{
  "data": {
    "memos": [
      {
        "id": "uuid",
        "title": "古いメモ",
        "deleted_at": "2026-03-10T00:00:00Z"
      }
    ],
    "todos": [...],
    "events": [...],
    "tools": [...],
    "tags": [...]
  }
}
```
