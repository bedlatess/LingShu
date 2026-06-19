ALTER TABLE users
    ADD COLUMN IF NOT EXISTS rpm_limit INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS concurrency_limit INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS token_revoked_at TIMESTAMPTZ;

INSERT INTO system_settings (key, value, description)
VALUES
    ('alert_enabled', 'false', '是否启用运营告警'),
    ('alert_channel_failure_threshold', '5', '渠道连续失败自动禁用阈值'),
    ('alert_gateway_5xx_rate_threshold', '0.20', '网关 5xx 告警比例阈值'),
    ('alert_upstream_error_rate_threshold', '0.20', '上游 401/429 告警比例阈值'),
    ('alert_low_balance_threshold', '5', '用户低余额告警阈值'),
    ('alert_email_recipients', '', '告警邮件收件人，多个用逗号分隔'),
    ('alert_webhook_url', '', '告警 Webhook 地址')
ON CONFLICT (key) DO NOTHING;

CREATE TABLE IF NOT EXISTS ops_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_key TEXT NOT NULL,
    severity TEXT NOT NULL DEFAULT 'warning',
    target_type TEXT NOT NULL DEFAULT '',
    target_id UUID,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'resolved')),
    fingerprint TEXT NOT NULL,
    last_notified_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ops_alerts_active_fingerprint
    ON ops_alerts(fingerprint)
    WHERE status='active';
