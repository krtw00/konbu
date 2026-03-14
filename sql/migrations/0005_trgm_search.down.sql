-- 0005_trgm_search.down.sql

DROP INDEX IF EXISTS idx_memos_title_trgm;
DROP INDEX IF EXISTS idx_memos_content_trgm;
DROP INDEX IF EXISTS idx_memo_rows_data_trgm;
DROP INDEX IF EXISTS idx_todos_title_trgm;
DROP INDEX IF EXISTS idx_calendar_events_title_trgm;

DROP EXTENSION IF EXISTS "pg_trgm";
