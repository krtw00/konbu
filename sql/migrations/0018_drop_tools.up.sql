-- Remove the tool launcher (bookmark) feature. KONBU-19.
DROP INDEX IF EXISTS idx_tools_url_trgm;
DROP INDEX IF EXISTS idx_tools_name_trgm;
DROP INDEX IF EXISTS idx_tools_user_id;
DROP TABLE IF EXISTS tools;
