CREATE INDEX IF NOT EXISTS idx_tools_name_trgm ON tools USING gin (name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_tools_url_trgm ON tools USING gin (url gin_trgm_ops);
