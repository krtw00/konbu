CREATE TABLE feedback_submissions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    email       TEXT NOT NULL,
    category    TEXT NOT NULL CHECK (category IN ('bug', 'feature', 'question', 'other')),
    message     TEXT NOT NULL,
    source_page TEXT NOT NULL DEFAULT '',
    user_agent  TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'closed')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_feedback_submissions_created_at ON feedback_submissions(created_at DESC);
CREATE INDEX idx_feedback_submissions_status ON feedback_submissions(status, created_at DESC);
