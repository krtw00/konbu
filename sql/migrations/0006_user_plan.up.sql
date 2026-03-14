-- 0006_user_plan.up.sql
-- Add plan column for Sponsor tier features.

ALTER TABLE users ADD COLUMN plan TEXT NOT NULL DEFAULT 'free';
