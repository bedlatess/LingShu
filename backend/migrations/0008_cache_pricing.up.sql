ALTER TABLE models
  ADD COLUMN cache_creation_price_per_1k NUMERIC(18,6) NOT NULL DEFAULT 0,
  ADD COLUMN cache_read_price_per_1k NUMERIC(18,6) NOT NULL DEFAULT 0;

ALTER TABLE gateway_requests
  ADD COLUMN cache_creation_tokens INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN cache_read_tokens INTEGER NOT NULL DEFAULT 0;
