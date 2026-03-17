-- konbu DDL
-- PostgreSQL 16+
--
-- 削除方針: 論理削除（deleted_at）をベースとする。
-- 物理削除はゴミ箱の期限切れ等、明示的な操作でのみ実行。
-- 全クエリは WHERE deleted_at IS NULL を基本とする。

-- =============================================================================
-- Extensions
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";   -- gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS "pg_trgm";    -- trigram full-text search

-- =============================================================================
-- Users
-- =============================================================================

-- ForwardAuth ヘッダーから自動登録。最初の登録ユーザーが管理者。
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL UNIQUE,
    name          TEXT NOT NULL DEFAULT '',
    password_hash TEXT,
    is_admin      BOOLEAN NOT NULL DEFAULT FALSE,
    plan          TEXT NOT NULL DEFAULT 'free',
    user_settings JSONB DEFAULT '{}'::jsonb,
    locale        TEXT DEFAULT 'en',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);

-- =============================================================================
-- API Keys
-- =============================================================================

-- CLI / bot 用 Bearer トークン。ユーザーごとに複数発行可。
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

-- メモ・ToDo・カレンダー横断。ユーザーごとに名前空間を分離。
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

-- type='markdown': content にマークダウン本文
-- type='table':    table_columns にカラム定義 JSONB、行データは memo_rows
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

-- テーブル型メモの行データ。行単位で追加・削除・ソート可能。
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
CREATE INDEX idx_memos_title_trgm ON memos USING gin (title gin_trgm_ops);
CREATE INDEX idx_memos_content_trgm ON memos USING gin (content gin_trgm_ops);
CREATE INDEX idx_memo_rows_memo_id ON memo_rows(memo_id, sort_order) WHERE deleted_at IS NULL;
CREATE INDEX idx_memo_rows_data_trgm ON memo_rows USING gin ((row_data::text) gin_trgm_ops);

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
CREATE INDEX idx_todos_title_trgm ON todos USING gin (title gin_trgm_ops);

-- =============================================================================
-- Calendars
-- =============================================================================

CREATE TABLE calendars (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL DEFAULT '',
    is_default  BOOLEAN NOT NULL DEFAULT FALSE,
    share_token TEXT UNIQUE,
    color       TEXT NOT NULL DEFAULT '#4F46E5',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_calendars_owner_id ON calendars(owner_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_calendars_default ON calendars(owner_id) WHERE is_default = TRUE AND deleted_at IS NULL;

CREATE TABLE calendar_members (
    calendar_id UUID NOT NULL REFERENCES calendars(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        TEXT NOT NULL DEFAULT 'editor' CHECK (role IN ('admin', 'editor', 'viewer')),
    color       TEXT NOT NULL DEFAULT '#4F46E5',
    joined_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (calendar_id, user_id)
);

CREATE INDEX idx_calendar_members_user_id ON calendar_members(user_id);

-- =============================================================================
-- Calendar Events
-- =============================================================================

CREATE TABLE calendar_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_by  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    calendar_id UUID REFERENCES calendars(id) ON DELETE SET NULL,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    start_at    TIMESTAMPTZ NOT NULL,
    end_at      TIMESTAMPTZ,
    all_day     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE TABLE calendar_event_tags (
    event_id UUID NOT NULL REFERENCES calendar_events(id) ON DELETE CASCADE,
    tag_id   UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (event_id, tag_id)
);

CREATE INDEX idx_calendar_events_user_id ON calendar_events(created_by) WHERE deleted_at IS NULL;
CREATE INDEX idx_calendar_events_range ON calendar_events(created_by, start_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_calendar_events_calendar_id ON calendar_events(calendar_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_calendar_events_title_trgm ON calendar_events USING gin (title gin_trgm_ops);

-- =============================================================================
-- Tools (launcher links)
-- =============================================================================

-- ユーザーごとのツールリンク。Web UI のランチャーカードに表示。
CREATE TABLE tools (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    url        TEXT NOT NULL,
    icon       TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_tools_user_id ON tools(user_id, sort_order) WHERE deleted_at IS NULL;
CREATE INDEX idx_tools_name_trgm ON tools USING gin (name gin_trgm_ops);
CREATE INDEX idx_tools_url_trgm ON tools USING gin (url gin_trgm_ops);

-- =============================================================================
-- Chat Sessions & Messages
-- =============================================================================

CREATE TABLE chat_sessions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title      TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_chat_sessions_user_id ON chat_sessions(user_id, updated_at DESC) WHERE deleted_at IS NULL;

CREATE TABLE chat_messages (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id   UUID NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    role         TEXT NOT NULL CHECK (role IN ('user', 'assistant', 'system', 'tool')),
    content      TEXT NOT NULL DEFAULT '',
    tool_calls   JSONB,
    tool_call_id TEXT,
    provider     TEXT,
    model        TEXT,
    input_tokens  INTEGER,
    output_tokens INTEGER,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_chat_messages_session_id ON chat_messages(session_id, created_at);
