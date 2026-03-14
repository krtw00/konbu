-- 0005_trgm_search.up.sql
-- Replace pg_bigm with pg_trgm for full-text search indexing.
-- pg_trgm is available on managed PostgreSQL services (Supabase, Neon, RDS, etc).

CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Drop pg_bigm indexes if they exist (from manual schema application)
DROP INDEX IF EXISTS idx_memos_title_bigm;
DROP INDEX IF EXISTS idx_memos_content_bigm;
DROP INDEX IF EXISTS idx_memo_rows_data_bigm;
DROP INDEX IF EXISTS idx_todos_title_bigm;
DROP INDEX IF EXISTS idx_calendar_events_title_bigm;

-- Create pg_trgm GIN indexes for ILIKE acceleration
CREATE INDEX idx_memos_title_trgm ON memos USING gin (title gin_trgm_ops);
CREATE INDEX idx_memos_content_trgm ON memos USING gin (content gin_trgm_ops);
CREATE INDEX idx_memo_rows_data_trgm ON memo_rows USING gin ((row_data::text) gin_trgm_ops);
CREATE INDEX idx_todos_title_trgm ON todos USING gin (title gin_trgm_ops);
CREATE INDEX idx_calendar_events_title_trgm ON calendar_events USING gin (title gin_trgm_ops);
