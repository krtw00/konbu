-- Recreate the calendar sharing / external integration schema (rollback of 0020).
-- KONBU-18. Restores calendars.share_token, calendar_members and
-- calendar_feed_tokens with their indexes.
ALTER TABLE calendars ADD COLUMN IF NOT EXISTS share_token TEXT UNIQUE;

CREATE TABLE IF NOT EXISTS calendar_members (
    calendar_id UUID NOT NULL REFERENCES calendars(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        TEXT NOT NULL DEFAULT 'editor' CHECK (role IN ('admin', 'editor', 'viewer')),
    color       TEXT NOT NULL DEFAULT '#4F46E5',
    joined_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (calendar_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_calendar_members_user_id ON calendar_members(user_id);

CREATE TABLE IF NOT EXISTS calendar_feed_tokens (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    token_hash   TEXT NOT NULL UNIQUE,
    last_used_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_calendar_feed_tokens_user_id
    ON calendar_feed_tokens(user_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_calendar_feed_tokens_token_hash
    ON calendar_feed_tokens(token_hash)
    WHERE deleted_at IS NULL;
