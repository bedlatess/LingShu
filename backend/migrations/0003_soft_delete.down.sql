DROP INDEX IF EXISTS idx_api_keys_alive;
DROP INDEX IF EXISTS idx_models_alive;
DROP INDEX IF EXISTS idx_channels_alive;

ALTER TABLE api_keys DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE models DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE upstream_channels DROP COLUMN IF EXISTS deleted_at;
