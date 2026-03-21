ALTER TABLE tools
ADD COLUMN icon_checked_at TIMESTAMPTZ;

UPDATE tools
SET icon_checked_at = created_at
WHERE url != '';
