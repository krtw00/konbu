-- Remove calendar sharing / external integration: multi-user calendars and the
-- outbound iCal feed. KONBU-18.
-- konbu is a personal-only digital system notebook; the calendar keeps its
-- per-user multiple-calendar model (owner-only), but every outbound / multi-user
-- path is removed:
--   - calendar_members (membership + admin/editor/viewer roles)
--   - calendars.share_token (invite share link)
--   - calendar_feed_tokens (outbound iCal feed tokens)
-- calendars and calendar_events themselves are preserved.
DROP INDEX IF EXISTS idx_calendar_members_user_id;
DROP TABLE IF EXISTS calendar_members;

DROP INDEX IF EXISTS idx_calendar_feed_tokens_token_hash;
DROP INDEX IF EXISTS idx_calendar_feed_tokens_user_id;
DROP TABLE IF EXISTS calendar_feed_tokens;

ALTER TABLE calendars DROP COLUMN IF EXISTS share_token;
