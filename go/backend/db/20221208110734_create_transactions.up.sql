CREATE TABLE IF NOT EXISTS transactions
(
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  wallet_id     UUID NOT NULL REFERENCES wallets(id),
  foreign_id    UUID NOT NULL,
  type          TEXT NOT NULL,
  state         TEXT NOT NULL,
  provider      TEXT NOT NULL,
  source        TEXT,
  destination   TEXT,
  note          TEXT,
  amount        BIGINT NOT NULL,
  asset_code    TEXT NOT NULL,
  asset_scale   INT NOT NULL,
  created_at    TIMESTAMP NOT NULL DEFAULT now(),
  updated_at    TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT  transactions_wallet_id_foreign_id UNIQUE (wallet_id, foreign_id),
  INDEX transactions_wallet_id_state(wallet_id, state)
);

CREATE TABLE IF NOT EXISTS transfers
(
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  transaction_id   UUID NOT NULL REFERENCES transactions(id),
  foreign_id       UUID NOT NULL,
  type             TEXT NOT NULL,
  state            TEXT NOT NULL,
  amount           BIGINT NOT NULL,
  asset_code       TEXT NOT NULL,
  asset_scale      INT NOT NULL,
  created_at       TIMESTAMP NOT NULL DEFAULT now(),
  updated_at       TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT transfers_transaction_id_foreign_id_type UNIQUE (transaction_id, foreign_id, type)
);