# #20 Google カレンダー取り込み表示 (iCal 購読)

> KONBU-20 / 骨子「一冊集約・inbound」。outbound feed 撤去 (KONBU-18) と対になり「全部流れ込む / 外に漏れない」を体現する。

## 概要

- **目的**: Google カレンダーの非公開 iCal URL を登録し、定期 fetch して konbu に read-only 表示する
- **スコープ (MVP)**: iCal URL 購読方式。既存の自前 iCal パーサ (`internal/service/import_service.go`) を再利用し、通知 sweep と同パターンの background loop で定期取得
- **非スコープ**: Google Calendar API / OAuth scope 拡張 (将来)。既存 Google OAuth はログイン用途のみで scope を増やさない。outbound (konbu → 外部) は持たない
- **方針**: 1 購読 = 1 専用 read-only カレンダー (ユーザー決定 2026-05-21)。read-only は「外部カレンダーに属する」で一意に決まる

## 要件

### 機能要件
- Google カレンダーの非公開 iCal URL を登録できる (専用カレンダーを自動生成)
- 登録した iCal を定期 fetch して konbu のカレンダーに表示
- 取り込みイベントは read-only (編集ダイアログを出さない) + 自前イベントとバッジ/色で区別
- 再 fetch で重複しない (VEVENT UID = `external_uid` で upsert dedup、上流から消えたものは削除)
- 購読の一覧・削除・手動 sync ができる
- owner-only (作成者のみ操作可)

### 非機能要件
- fetch は timeout + サイズ上限付き。URL は https 必須
- 購読ゼロでも loop は安全に no-op

## 設計

### DB マイグレーション (0021)

`calendars` 拡張:
```sql
ALTER TABLE calendars ADD COLUMN is_external boolean NOT NULL DEFAULT false;
```

`calendar_events` 拡張 (dedup キー):
```sql
ALTER TABLE calendar_events ADD COLUMN external_uid text;
CREATE UNIQUE INDEX idx_calendar_events_external
  ON calendar_events (calendar_id, external_uid)
  WHERE external_uid IS NOT NULL AND deleted_at IS NULL;
```

新規 `calendar_subscriptions`:
```sql
CREATE TABLE calendar_subscriptions (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id        uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    calendar_id     uuid NOT NULL REFERENCES calendars(id) ON DELETE CASCADE,
    ical_url        text NOT NULL,
    last_fetched_at timestamptz,
    last_error      text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz
);
CREATE UNIQUE INDEX idx_calendar_subscriptions_calendar ON calendar_subscriptions (calendar_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_calendar_subscriptions_owner ON calendar_subscriptions (owner_id) WHERE deleted_at IS NULL;
```

down マイグレーションは DROP TABLE / DROP COLUMN。

### API (owner-only, `/calendars/subscriptions`)

| メソッド | パス | 説明 |
|---|---|---|
| GET | `/calendars/subscriptions` | 購読一覧 (last_fetched_at / last_error 含む) |
| POST | `/calendars/subscriptions` | 登録 (`{name, ical_url, color}`) → 専用カレンダー + 購読作成 → 初回 fetch |
| DELETE | `/calendars/subscriptions/:id` | 購読 + 専用カレンダー + 取り込みイベントを削除 |
| POST | `/calendars/subscriptions/:id/sync` | 手動 fetch |

### サービス層

- `subscription_service.go` (新規):
  - `Create`: https URL 検証 → `is_external=true` のカレンダー作成 → 購読レコード作成 → `syncOne` 初回実行
  - `syncOne(subscription)`: GET (timeout 15s / 上限 5MB) → `parseICal` 再利用 → fetched UID 集合を作り、`calendar_id` 配下を external_uid で upsert、集合に無い既存 external イベントを論理削除 → `last_fetched_at`/`last_error` 更新
  - `SyncAll`: 全購読を順次 syncOne (loop から呼ぶ)
- `parseICal` を import_service と共有 (現状 private なら export または共通関数に切り出し)

### 定期実行 loop

- 通知 sweep (`notification_service.StartLoop`) と同じ goroutine + ticker パターンで `subscription_service.SyncAll` を回す
- 間隔は `CALENDAR_SYNC_INTERVAL` (Go duration, default `30m`)
- `cmd/server/main.go` の起動時に開始

### リポジトリ層

- `calendar_subscriptions.go` (新規): CRUD
- `events.go` 拡張: external_uid 付き upsert / 指定 calendar の external イベント一括取得・stale 削除
- `calendars.go`: is_external 反映

### フロントエンド

- `types/api.ts`: `Calendar` に `is_external`、`CalendarEvent` に `external_uid?`、`CalendarSubscription` 型追加
- `lib/api.ts`: `listSubscriptions` / `createSubscription` / `deleteSubscription` / `syncSubscription`
- 購読管理 UI (SettingsPage か CalendarPage のカレンダー管理ダイアログ内): URL + 名前 + 色で登録、一覧 (最終取得日時 / エラー表示)、削除、手動 sync
- `CalendarPage.tsx`: `is_external` カレンダーのイベントは編集ダイアログを開かない + バッジ表示
- i18n (en/ja)

## タスク分解

### バックエンド
- [ ] migration 0021 (calendars.is_external / calendar_events.external_uid / calendar_subscriptions)
- [ ] schema.sql 反映
- [ ] model: CalendarSubscription、Calendar.IsExternal、CalendarEvent.ExternalUID
- [ ] repository: calendar_subscriptions CRUD、events の external upsert / stale 削除
- [ ] service: subscription_service (Create / syncOne / SyncAll)、parseICal 共有化
- [ ] handler: subscription routes (owner-only)
- [ ] main.go: ルート mount + sync loop 起動、config に CALENDAR_SYNC_INTERVAL
- [ ] go build / go vet / go test

### フロントエンド
- [ ] types / api client
- [ ] 購読管理 UI (登録 / 一覧 / 削除 / 手動 sync)
- [ ] CalendarPage: external イベント read-only + バッジ
- [ ] i18n (en/ja)
- [ ] tsc / vitest

## リスク・検討
- **SSRF**: サーバーが任意 URL を fetch する。MVP は個人専用・単一ユーザーかつ自分の URL のため低リスク。https 必須 + timeout + サイズ上限で最小限ガード。private IP block 等の本格対策は将来
- **RRULE / timezone**: 既存パーサは FREQ のみ・UTC 扱い。Google の複雑な RRULE / TZID は MVP では簡易対応に留める (既存 import と同等)
- **iCal URL = 秘密情報**: token を含む。MVP は平文保存 (他 token と同等)。暗号化は将来検討
- **stale 削除の単位**: 専用カレンダー配下の external_uid 付きイベントのみ対象 (自前イベントは触らない)
