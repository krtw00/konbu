CREATE TABLE IF NOT EXISTS notification_sent_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    resource_type TEXT NOT NULL,
    resource_id UUID NOT NULL,
    kind TEXT NOT NULL,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, resource_type, resource_id, kind)
);

CREATE INDEX IF NOT EXISTS idx_notification_sent_log_user_id
    ON notification_sent_log (user_id);
