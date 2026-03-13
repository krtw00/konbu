# Database

## 概要

PostgreSQL 16+を使用。全文検索にはpg_bigm拡張（2-gram、日本語対応）を利用。

## ER図

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

## テーブル一覧

| テーブル | 説明 |
|---------|------|
| users | ユーザー (メール+パスワード) |
| api_keys | APIキー (Bearer認証) |
| tags | タグ (ユーザーごと、(user_id, name) ユニーク) |
| memos | メモ (Markdown / テーブル型) |
| memo_tags | メモ↔タグ |
| memo_rows | テーブル型メモの行データ (JSONB) |
| todos | ToDo (status: open/done) |
| todo_tags | ToDo↔タグ |
| calendar_events | カレンダー予定 (終日/時間指定、繰り返し) |
| calendar_event_tags | 予定↔タグ |
| tools | ツールランチャー (カテゴリ分類) |

完全なDDLは [schema.sql](schema.sql) を参照。

## 論理削除

全テーブルに `deleted_at TIMESTAMPTZ` カラムを持つ。

- DELETE APIは `deleted_at = now()` をセット
- 全SELECTクエリに `WHERE deleted_at IS NULL` を付与
- 物理削除は未実装（将来のゴミ箱機能で対応予定）

## タグ

- メモ・ToDo・予定の3リソースで共有
- 作成リクエストで存在しないタグ名を指定すると自動作成（暗黙的upsert）
- 中間テーブル（memo_tags, todo_tags, calendar_event_tags）で多対多
- service層でタグの付け替え（全削除→再挿入）を処理

## マイグレーション

`sql/migrations/` 配下にup/downペアで管理。

| ファイル | 内容 |
|---------|------|
| 0001_initial | 全テーブル作成 |
| 0002_auth_password | users に password_hash, user_settings, locale 追加 |
| 0003_recurring_events | calendar_events に recurrence_rule, recurrence_end 追加 |
| 0004_tool_category | tools に category 追加 |

Docker Composeでは `docker-entrypoint-initdb.d` にマウントして初期化。ネイティブ環境では `psql` で順に実行。

## 検索

pg_bigmを使った2-gram全文検索。

- メモ: title + content
- ToDo: title + description
- 予定: title + description

`ILIKE` による部分一致検索をベースに、pg_bigmのGINインデックスで高速化。日本語の分かち書きなしで動作する。
