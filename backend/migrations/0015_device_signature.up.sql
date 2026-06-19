INSERT INTO system_settings (key, value, description)
VALUES (
    'device_secret_key',
    encode(gen_random_bytes(32), 'hex'),
    'Public device signature key used by the browser to bind device id and user agent'
)
ON CONFLICT (key) DO NOTHING;
