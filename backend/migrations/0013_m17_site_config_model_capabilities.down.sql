ALTER TABLE models
    DROP COLUMN IF EXISTS supports_stream,
    DROP COLUMN IF EXISTS supports_tools,
    DROP COLUMN IF EXISTS supports_vision;

DELETE FROM system_settings WHERE key IN ('api_base_url', 'alert_webhook_provider', 'trusted_proxy_enabled', 'trusted_proxy_hops');
