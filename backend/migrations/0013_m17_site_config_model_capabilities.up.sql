ALTER TABLE models
    ADD COLUMN IF NOT EXISTS supports_stream BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS supports_tools BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS supports_vision BOOLEAN NOT NULL DEFAULT false;

UPDATE models
SET supports_stream = CASE WHEN type = 'chat' THEN true ELSE supports_stream END,
    supports_tools = CASE WHEN type = 'chat' THEN true ELSE supports_tools END,
    supports_vision = CASE WHEN type IN ('image', 'video') THEN true ELSE supports_vision END
WHERE deleted_at IS NULL;

INSERT INTO system_settings (key, value, description)
VALUES
    ('api_base_url', '', 'Public API base_url shown in integration docs; empty means browser origin + /v1'),
    ('alert_webhook_provider', 'generic', 'Alert webhook provider: generic, wechat, feishu, dingtalk, or discord'),
    ('trusted_proxy_enabled', 'false', 'Whether to trust reverse proxy IP headers such as X-Forwarded-For'),
    ('trusted_proxy_hops', '1', 'Trusted reverse proxy hop count for X-Forwarded-For')
ON CONFLICT (key) DO NOTHING;
