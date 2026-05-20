-- Recreate the tools table and its indexes (rollback of 0018). KONBU-19.
CREATE TABLE tools (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    url             TEXT NOT NULL,
    icon            TEXT NOT NULL DEFAULT '',
    icon_checked_at TIMESTAMPTZ,
    category        TEXT NOT NULL DEFAULT '',
    sort_order      INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_tools_user_id ON tools(user_id, sort_order) WHERE deleted_at IS NULL;
CREATE INDEX idx_tools_name_trgm ON tools USING gin (name gin_trgm_ops);
CREATE INDEX idx_tools_url_trgm ON tools USING gin (url gin_trgm_ops);
