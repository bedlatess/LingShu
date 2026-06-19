ALTER TABLE users
    ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT false;

INSERT INTO system_settings (key, value, description)
VALUES
    ('site_name', 'LingShu', 'Public site name'),
    ('registration_mode', 'closed', 'Registration mode: open, invite, or closed'),
    ('smtp_host', '', 'SMTP server host'),
    ('smtp_port', '587', 'SMTP server port'),
    ('smtp_user', '', 'SMTP username'),
    ('smtp_pass', '', 'SMTP password'),
    ('smtp_from', '', 'SMTP sender address'),
    ('smtp_tls', 'true', 'Whether SMTP uses TLS or STARTTLS'),
    ('captcha_enabled', 'false', 'Whether captcha is required for login/register'),
    ('captcha_provider', '', 'Captcha provider name'),
    ('site_logo_url', '', 'Site logo URL'),
    ('site_icp', '', 'ICP filing number'),
    ('site_police_beian', '', 'Public security filing number'),
    ('tos_url', '/legal/tos', 'Terms of service URL'),
    ('privacy_url', '/legal/privacy', 'Privacy policy URL'),
    ('contact_email', '', 'Public contact email'),
    ('brand_primary_color', '', 'Optional brand primary color'),
    ('legal_tos_markdown', '# 服务条款\n\n管理员尚未配置服务条款正文。', 'Terms of service markdown'),
    ('legal_privacy_markdown', '# 隐私政策\n\n管理员尚未配置隐私政策正文。', 'Privacy policy markdown')
ON CONFLICT (key) DO NOTHING;
