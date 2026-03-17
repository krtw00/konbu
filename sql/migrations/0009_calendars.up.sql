-- 0009_calendars.up.sql
-- グループカレンダー: TimeTree型のカレンダー単位でイベントを管理

CREATE TABLE calendars (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT NOT NULL DEFAULT '',
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    token      TEXT UNIQUE,
    color      TEXT NOT NULL DEFAULT '#3B82F6',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_calendars_owner ON calendars(owner_id) WHERE deleted_at IS NULL;

CREATE TABLE calendar_members (
    calendar_id UUID NOT NULL REFERENCES calendars(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        TEXT NOT NULL DEFAULT 'member' CHECK (role IN ('admin', 'member', 'viewer')),
    color       TEXT NOT NULL DEFAULT '',
    joined_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (calendar_id, user_id)
);

-- 既存ユーザーごとにデフォルトカレンダーを作成
INSERT INTO calendars (id, owner_id, name, is_default)
SELECT gen_random_uuid(), id, 'My Calendar', TRUE
FROM users WHERE deleted_at IS NULL;

-- オーナーを admin メンバーとして追加
INSERT INTO calendar_members (calendar_id, user_id, role)
SELECT c.id, c.owner_id, 'admin'
FROM calendars c WHERE c.is_default = TRUE;

-- calendar_events に calendar_id を追加し、既存イベントを紐づけ
ALTER TABLE calendar_events ADD COLUMN calendar_id UUID REFERENCES calendars(id);

UPDATE calendar_events e
SET calendar_id = c.id
FROM calendars c
WHERE c.owner_id = e.user_id AND c.is_default = TRUE;

ALTER TABLE calendar_events ALTER COLUMN calendar_id SET NOT NULL;

-- user_id を created_by に改名
ALTER TABLE calendar_events RENAME COLUMN user_id TO created_by;

CREATE INDEX idx_calendar_events_calendar ON calendar_events(calendar_id) WHERE deleted_at IS NULL;
