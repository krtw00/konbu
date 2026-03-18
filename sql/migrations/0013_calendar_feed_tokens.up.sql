CREATE TABLE calendar_feed_tokens (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    token_hash   TEXT NOT NULL UNIQUE,
    last_used_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX idx_calendar_feed_tokens_user_id
    ON calendar_feed_tokens(user_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_calendar_feed_tokens_token_hash
    ON calendar_feed_tokens(token_hash)
    WHERE deleted_at IS NULL;
