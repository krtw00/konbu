-- Remove the public sharing features: PublicShare and PublishedResource. KONBU-17.
-- konbu is a personal-only digital system notebook; external publish/share links
-- are removed to commit fully to the single-user model.
DROP INDEX IF EXISTS idx_published_resources_visibility;
DROP TABLE IF EXISTS published_resources;

DROP INDEX IF EXISTS idx_public_shares_token;
DROP TABLE IF EXISTS public_shares;
