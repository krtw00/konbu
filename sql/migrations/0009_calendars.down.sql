-- 0009_calendars.down.sql

DROP INDEX IF EXISTS idx_calendar_events_calendar;

ALTER TABLE calendar_events RENAME COLUMN created_by TO user_id;
ALTER TABLE calendar_events DROP COLUMN IF EXISTS calendar_id;

DROP TABLE IF EXISTS calendar_members;
DROP TABLE IF EXISTS calendars;
