ALTER TABLE upstream_channels
DROP CONSTRAINT IF EXISTS upstream_channels_provider_type_check;

ALTER TABLE upstream_channels
ADD CONSTRAINT upstream_channels_provider_type_check
CHECK (provider_type IN ('openai', 'claude', 'gemini', 'custom'));
