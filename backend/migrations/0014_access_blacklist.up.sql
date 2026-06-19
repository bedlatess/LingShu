CREATE TABLE IF NOT EXISTS access_blacklist (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kind TEXT NOT NULL CHECK (kind IN ('ip', 'cidr', 'device')),
    value TEXT NOT NULL,
    scope TEXT NOT NULL DEFAULT 'all' CHECK (scope IN ('login', 'gateway', 'all')),
    reason TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT 'manual' CHECK (source IN ('manual', 'auto')),
    active BOOLEAN NOT NULL DEFAULT true,
    created_by UUID REFERENCES users(id),
    released_by UUID REFERENCES users(id),
    expires_at TIMESTAMPTZ,
    released_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_access_blacklist_active_kind_value_scope
    ON access_blacklist(kind, value, scope)
    WHERE active = true;

CREATE INDEX IF NOT EXISTS idx_access_blacklist_active_expires
    ON access_blacklist(active, expires_at);

INSERT INTO system_settings (key, value, description)
VALUES
    ('access_blacklist_auto_enabled', 'true', '是否启用登录失败自动拉黑'),
    ('access_blacklist_login_fail_threshold', '10', '同一 IP 或设备登录失败自动拉黑阈值'),
    ('access_blacklist_auto_ttl_days', '7', '自动拉黑默认有效天数')
ON CONFLICT (key) DO NOTHING;
