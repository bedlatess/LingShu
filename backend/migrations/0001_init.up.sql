CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT NOT NULL UNIQUE,
    email TEXT UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('admin', 'user')),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'banned')),
    balance NUMERIC(18,6) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_login_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_prefix TEXT NOT NULL,
    key_hash TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'disabled')),
    rpm_limit INTEGER NOT NULL DEFAULT 60,
    concurrency_limit INTEGER NOT NULL DEFAULT 5,
    expires_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    public_name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('chat', 'embedding', 'image', 'video')),
    model_group TEXT NOT NULL DEFAULT '',
    billing_mode TEXT NOT NULL CHECK (billing_mode IN ('token', 'per_call')),
    input_price_per_1k NUMERIC(18,6) NOT NULL DEFAULT 0,
    output_price_per_1k NUMERIC(18,6) NOT NULL DEFAULT 0,
    price_per_call NUMERIC(18,6) NOT NULL DEFAULT 0,
    rate_multiplier NUMERIC(6,3) NOT NULL DEFAULT 1.200,
    status TEXT NOT NULL DEFAULT 'enabled' CHECK (status IN ('enabled', 'disabled')),
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS upstream_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    provider_type TEXT NOT NULL CHECK (provider_type IN ('openai', 'claude', 'gemini', 'custom')),
    base_url TEXT NOT NULL,
    api_key_encrypted TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'enabled' CHECK (status IN ('enabled', 'disabled')),
    weight INTEGER NOT NULL DEFAULT 1 CHECK (weight > 0),
    timeout_seconds INTEGER NOT NULL DEFAULT 120,
    rpm_limit INTEGER NOT NULL DEFAULT 60,
    concurrency_limit INTEGER NOT NULL DEFAULT 5,
    fail_threshold INTEGER NOT NULL DEFAULT 5,
    fail_count INTEGER NOT NULL DEFAULT 0,
    health TEXT NOT NULL DEFAULT 'healthy' CHECK (health IN ('healthy', 'unhealthy')),
    last_success_at TIMESTAMPTZ,
    last_error_at TIMESTAMPTZ,
    last_error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS channel_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES upstream_channels(id) ON DELETE CASCADE,
    model_id UUID NOT NULL REFERENCES models(id) ON DELETE CASCADE,
    upstream_model_name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'enabled' CHECK (status IN ('enabled', 'disabled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (channel_id, model_id)
);

CREATE TABLE IF NOT EXISTS gateway_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id TEXT NOT NULL UNIQUE,
    user_id UUID NOT NULL REFERENCES users(id),
    api_key_id UUID NOT NULL REFERENCES api_keys(id),
    model_id UUID REFERENCES models(id),
    channel_id UUID REFERENCES upstream_channels(id),
    endpoint TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('success', 'failed', 'partial')),
    http_status INTEGER NOT NULL,
    prompt_tokens INTEGER NOT NULL DEFAULT 0,
    completion_tokens INTEGER NOT NULL DEFAULT 0,
    total_tokens INTEGER NOT NULL DEFAULT 0,
    base_cost NUMERIC(18,6) NOT NULL DEFAULT 0,
    rate_multiplier NUMERIC(6,3) NOT NULL DEFAULT 1.000,
    charge NUMERIC(18,6) NOT NULL DEFAULT 0,
    is_stream BOOLEAN NOT NULL DEFAULT false,
    is_estimated BOOLEAN NOT NULL DEFAULT false,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    error_code TEXT,
    error_message TEXT,
    client_ip INET,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS balance_ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    type TEXT NOT NULL CHECK (type IN ('admin_grant', 'admin_deduct', 'redeem', 'usage_charge', 'refund', 'adjustment')),
    amount NUMERIC(18,6) NOT NULL,
    balance_before NUMERIC(18,6) NOT NULL,
    balance_after NUMERIC(18,6) NOT NULL,
    base_cost NUMERIC(18,6),
    rate_multiplier NUMERIC(6,3),
    related_type TEXT,
    related_id UUID,
    operator_id UUID REFERENCES users(id),
    remark TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS redeem_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code_hash TEXT NOT NULL UNIQUE,
    code_prefix TEXT NOT NULL,
    batch_name TEXT NOT NULL DEFAULT '',
    amount NUMERIC(18,6) NOT NULL,
    status TEXT NOT NULL DEFAULT 'unused' CHECK (status IN ('unused', 'used', 'expired', 'disabled')),
    max_uses INTEGER NOT NULL DEFAULT 1,
    used_count INTEGER NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS redeem_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    redeem_code_id UUID NOT NULL REFERENCES redeem_codes(id),
    user_id UUID NOT NULL REFERENCES users(id),
    amount NUMERIC(18,6) NOT NULL,
    ledger_id UUID REFERENCES balance_ledger(id),
    client_ip INET,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS announcements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'offline' CHECK (status IN ('online', 'offline')),
    priority INTEGER NOT NULL DEFAULT 0,
    pinned BOOLEAN NOT NULL DEFAULT false,
    publish_at TIMESTAMPTZ,
    expire_at TIMESTAMPTZ,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id UUID REFERENCES users(id),
    action TEXT NOT NULL,
    target_type TEXT NOT NULL,
    target_id UUID,
    before_snapshot JSONB,
    after_snapshot JSONB,
    ip INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS system_settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    updated_by UUID REFERENCES users(id),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_channel_models_model_id ON channel_models(model_id);
CREATE INDEX IF NOT EXISTS idx_gateway_requests_user_created ON gateway_requests(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_gateway_requests_model_created ON gateway_requests(model_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_balance_ledger_user_created ON balance_ledger(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_redeem_records_user_created ON redeem_records(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_created ON audit_logs(actor_id, created_at DESC);

INSERT INTO system_settings (key, value, description)
VALUES
    ('site_name', 'LingShu', 'Public site name'),
    ('registration_enabled', 'false', 'Whether users can register themselves'),
    ('default_rate_multiplier', '1.2', 'Default model rate multiplier'),
    ('api_key_prefix', 'lsk_live_', 'Platform API key prefix'),
    ('sticky_session_enabled', 'true', 'Prefer stable upstream channel for the same session'),
    ('max_retry', '2', 'Maximum upstream retry attempts'),
    ('default_rpm_limit', '60', 'Default per-key RPM limit'),
    ('default_concurrency_limit', '5', 'Default per-key concurrency limit'),
    ('default_user_balance', '0', 'Initial balance for new users'),
    ('contact_info', '', 'Support contact information')
ON CONFLICT (key) DO NOTHING;
