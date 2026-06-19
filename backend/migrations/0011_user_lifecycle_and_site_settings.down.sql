ALTER TABLE users
    DROP COLUMN IF EXISTS email_verified;

DELETE FROM system_settings
WHERE key IN (
    'registration_mode',
    'smtp_host',
    'smtp_port',
    'smtp_user',
    'smtp_pass',
    'smtp_from',
    'smtp_tls',
    'captcha_enabled',
    'captcha_provider',
    'site_logo_url',
    'site_icp',
    'site_police_beian',
    'tos_url',
    'privacy_url',
    'contact_email',
    'brand_primary_color',
    'legal_tos_markdown',
    'legal_privacy_markdown'
);

