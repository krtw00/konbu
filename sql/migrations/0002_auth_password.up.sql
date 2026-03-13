ALTER TABLE users ADD COLUMN password_hash TEXT;
ALTER TABLE users ADD COLUMN user_settings JSONB DEFAULT '{}'::jsonb;
ALTER TABLE users ADD COLUMN locale TEXT DEFAULT 'en';
