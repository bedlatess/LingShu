DELETE FROM system_settings
WHERE key IN (
    'access_blacklist_auto_enabled',
    'access_blacklist_login_fail_threshold',
    'access_blacklist_auto_ttl_days'
);

DROP TABLE IF EXISTS access_blacklist;
