-- KONBU-20. Google カレンダー取り込み (iCal 購読) のための schema 拡張。
-- 1 購読 = 1 専用 read-only カレンダー (is_external=true)。

ALTER TABLE calendars ADD COLUMN IF NOT EXISTS is_external boolean NOT NULL DEFAULT false;

ALTER TABLE calendar_events ADD COLUMN IF NOT EXISTS external_uid text;

CREATE UNIQUE INDEX IF NOT EXISTS idx_calendar_events_external
    ON calendar_events (calendar_id, external_uid)
    WHERE external_uid IS NOT NULL AND deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS calendar_subscriptions (
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

CREATE UNIQUE INDEX IF NOT EXISTS idx_calendar_subscriptions_calendar
    ON calendar_subscriptions (calendar_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_calendar_subscriptions_owner
    ON calendar_subscriptions (owner_id) WHERE deleted_at IS NULL;
