ALTER TABLE gateway_requests
  ADD COLUMN first_token_ms INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN upstream_model_name TEXT,
  ADD COLUMN image_output_tokens INTEGER NOT NULL DEFAULT 0;
