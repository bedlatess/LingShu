ALTER TABLE upstream_channels ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE models ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_channels_alive ON upstream_channels(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_models_alive ON models(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_api_keys_alive ON api_keys(deleted_at) WHERE deleted_at IS NULL;
