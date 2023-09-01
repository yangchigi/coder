-- Drop the table first, as this will fail if
-- there are any remaining foreign key references.
DROP TABLE IF EXISTS dbcrypt_keys;

-- If that worked, we can drop the columns.
ALTER TABLE git_auth_links
DROP COLUMN IF EXISTS oauth_access_token_key_id,
DROP COLUMN IF EXISTS oauth_refresh_token_key_id;

ALTER TABLE user_links
DROP COLUMN IF EXISTS oauth_access_token_key_id,
DROP COLUMN IF EXISTS oauth_refresh_token_key_id;

