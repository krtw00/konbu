-- Rollback of 0021 (KONBU-20).

DROP INDEX IF EXISTS idx_calendar_subscriptions_owner;
DROP INDEX IF EXISTS idx_calendar_subscriptions_calendar;
DROP TABLE IF EXISTS calendar_subscriptions;

DROP INDEX IF EXISTS idx_calendar_events_external;
ALTER TABLE calendar_events DROP COLUMN IF EXISTS external_uid;

ALTER TABLE calendars DROP COLUMN IF EXISTS is_external;
