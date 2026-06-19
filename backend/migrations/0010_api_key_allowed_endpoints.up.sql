ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS allowed_endpoints TEXT[] NOT NULL DEFAULT '{}';

