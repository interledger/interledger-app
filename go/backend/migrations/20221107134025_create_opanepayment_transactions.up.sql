CREATE TABLE IF NOT EXISTS openpayments_transactions
(
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  wallet_id     UUID NOT NULL REFERENCES wallets(id),
  foreign_id    UUID NOT NULL,
  foreign_type  TEXT NOT NULL,
  state         INT NOT NULL,
  source        TEXT,
  destination   TEXT,
  note          TEXT,
  amount        BIGINT NOT NULL,
  asset_code    TEXT NOT NULL,
  asset_scale   INT NOT NULL,
  created_at    TIMESTAMP NOT NULL DEFAULT now(),
  updated_at    TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT  openpayments_transactions_foreign_id UNIQUE (foreign_id),
  INDEX openpayments_transactions_wallet_id_state(wallet_id, state)
);

ALTER TABLE openpayments_outgoing_payment ADD COLUMN completed BOOLEAN NOT NULL DEFAULT FALSE;