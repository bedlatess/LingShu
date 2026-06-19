DROP TABLE IF EXISTS ops_alerts;

ALTER TABLE users
    DROP COLUMN IF EXISTS token_revoked_at,
    DROP COLUMN IF EXISTS concurrency_limit,
    DROP COLUMN IF EXISTS rpm_limit;

DELETE FROM system_settings
WHERE key IN (
    'alert_enabled',
    'alert_channel_failure_threshold',
    'alert_gateway_5xx_rate_threshold',
    'alert_upstream_error_rate_threshold',
    'alert_low_balance_threshold',
    'alert_email_recipients',
    'alert_webhook_url'
);
