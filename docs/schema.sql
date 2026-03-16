-- konbu DDL（参照用 — 全マイグレーション統合版）
-- PostgreSQL 16+
-- 更新: 2026-03-13
--
-- 実際のDB構築は sql/migrations/ 配下のマイグレーションファイルを使用すること。
-- このファイルは現在のスキーマの全体像を把握するための参照用。

-- =============================================================================
-- Extensions
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";   -- gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS "pg_trgm";    -- trigram full-text search

-- =============================================================================
-- Users
-- =============================================================================

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL UNIQUE,
    name          TEXT NOT NULL DEFAULT '',
    password_hash TEXT,                                          -- 0002
    user_settings JSONB DEFAULT '{}'::jsonb,                     -- 0002
    locale        TEXT DEFAULT 'en',                             -- 0002
    is_admin      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);

-- =============================================================================
-- API Keys
-- =============================================================================

CREATE TABLE api_keys (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name         TEXT NOT NULL DEFAULT '',
    key_hash     TEXT NOT NULL,
    last_used_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX idx_api_keys_user_id ON api_keys(user_id) WHERE deleted_at IS NULL;

-- =============================================================================
-- Tags
-- =============================================================================

CREATE TABLE tags (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (user_id, name)
);

CREATE INDEX idx_tags_user_id ON tags(user_id) WHERE deleted_at IS NULL;

-- =============================================================================
-- Memos
-- =============================================================================

CREATE TABLE memos (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title         TEXT NOT NULL DEFAULT '',
    type          TEXT NOT NULL DEFAULT 'markdown' CHECK (type IN ('markdown', 'table')),
    content       TEXT,
    table_columns JSONB,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);

CREATE TABLE memo_tags (
    memo_id UUID NOT NULL REFERENCES memos(id) ON DELETE CASCADE,
    tag_id  UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (memo_id, tag_id)
);

CREATE TABLE memo_rows (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    memo_id    UUID NOT NULL REFERENCES memos(id) ON DELETE CASCADE,
    row_data   JSONB NOT NULL DEFAULT '{}',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_memos_user_id ON memos(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_memos_user_type ON memos(user_id, type) WHERE deleted_at IS NULL;
CREATE INDEX idx_memo_rows_memo_id ON memo_rows(memo_id, sort_order) WHERE deleted_at IS NULL;

-- =============================================================================
-- ToDos
-- =============================================================================

CREATE TABLE todos (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'done')),
    due_date    DATE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE TABLE todo_tags (
    todo_id UUID NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
    tag_id  UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (todo_id, tag_id)
);

CREATE INDEX idx_todos_user_id ON todos(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_todos_user_status ON todos(user_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_todos_due_date ON todos(user_id, due_date) WHERE deleted_at IS NULL;

-- =============================================================================
-- Calendar Events
-- =============================================================================

CREATE TABLE calendar_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title           TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    start_at        TIMESTAMPTZ NOT NULL,
    end_at          TIMESTAMPTZ,
    all_day         BOOLEAN NOT NULL DEFAULT FALSE,
    recurrence_rule TEXT,                                        -- 0003
    recurrence_end  DATE,                                        -- 0003
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE TABLE calendar_event_tags (
    event_id UUID NOT NULL REFERENCES calendar_events(id) ON DELETE CASCADE,
    tag_id   UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (event_id, tag_id)
);

CREATE INDEX idx_calendar_events_user_id ON calendar_events(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_calendar_events_range ON calendar_events(user_id, start_at) WHERE deleted_at IS NULL;

-- =============================================================================
-- Tools
-- =============================================================================

CREATE TABLE tools (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    url        TEXT NOT NULL,
    icon       TEXT NOT NULL DEFAULT '',
    category   TEXT NOT NULL DEFAULT '',                          -- 0004
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_tools_user_id ON tools(user_id, sort_order) WHERE deleted_at IS NULL;
CREATE INDEX idx_tools_name_trgm ON tools USING gin (name gin_trgm_ops);
CREATE INDEX idx_tools_url_trgm ON tools USING gin (url gin_trgm_ops);
