ALTER TABLE gateway_requests
  DROP COLUMN IF EXISTS cache_read_tokens,
  DROP COLUMN IF EXISTS cache_creation_tokens;

ALTER TABLE models
  DROP COLUMN IF EXISTS cache_read_price_per_1k,
  DROP COLUMN IF EXISTS cache_creation_price_per_1k;
