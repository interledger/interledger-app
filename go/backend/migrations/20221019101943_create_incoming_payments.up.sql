CREATE TABLE IF NOT EXISTS openpayments_incoming_payment
(
  id                  UUID PRIMARY KEY   DEFAULT gen_random_uuid(),
  payment_pointer_id  UUID REFERENCES payment_pointers,
  --quote_id UUID REFERENCES openpayments_quotes,
  asset_code          TEXT NOT NULL,
  asset_scale         INT NOT NULL,
  incoming_amount     BIGINT NOT NULL,
  received_amount     BIGINT NOT NULL,
  completed           BOOLEAN DEFAULT FALSE,
  expires_at          TIMESTAMP NOT NULL,
  external_ref        TEXT,
  ilp_stream_id       TEXT,
  ilp_address         TEXT,
  ilp_shared_secret   TEXT,
  created_at          TIMESTAMP NOT NULL DEFAULT now(),
  updated_at          TIMESTAMP NOT NULL DEFAULT now()
)