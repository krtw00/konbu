-- 0010_fix_calendars_schema.up.sql
-- Fix column name mismatch: token → share_token, add missing updated_at, fix role/color defaults

-- Rename token to share_token if token column exists
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'calendars' AND column_name = 'token') THEN
        ALTER TABLE calendars RENAME COLUMN token TO share_token;
    END IF;
END $$;

-- Add updated_at if missing
ALTER TABLE calendars ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- Fix role check constraint: allow 'editor' in addition to existing values
ALTER TABLE calendar_members DROP CONSTRAINT IF EXISTS calendar_members_role_check;
ALTER TABLE calendar_members ADD CONSTRAINT calendar_members_role_check CHECK (role IN ('admin', 'editor', 'member', 'viewer'));

-- Update existing 'member' roles to 'editor' for consistency
UPDATE calendar_members SET role = 'editor' WHERE role = 'member';

-- Add index on calendar_members.user_id if missing
CREATE INDEX IF NOT EXISTS idx_calendar_members_user_id ON calendar_members(user_id);

-- Add unique index on default calendar if missing
CREATE UNIQUE INDEX IF NOT EXISTS idx_calendars_default ON calendars(owner_id) WHERE is_default = TRUE AND deleted_at IS NULL;
