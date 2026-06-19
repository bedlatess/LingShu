ALTER TABLE gateway_requests
  DROP COLUMN IF EXISTS image_output_tokens,
  DROP COLUMN IF EXISTS upstream_model_name,
  DROP COLUMN IF EXISTS first_token_ms;
