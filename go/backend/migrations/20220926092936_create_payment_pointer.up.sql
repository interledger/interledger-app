CREATE TABLE IF NOT EXISTS payment_pointers (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  wallet_id   UUID NOT NULL,
  url         TEXT NOT NULL,
  alias       TEXT,
  asset       TEXT NOT NULL,
  scale       INT NOT NULL,
  created_at  TIMESTAMP NOT NULL DEFAULT now(),
  updated_at  TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT fk_payment_pointer_wallets
    FOREIGN KEY(wallet_id) REFERENCES wallets(id),
  CONSTRAINT  payment_pointers_url UNIQUE (url),
  INDEX wallet_id_ind(wallet_id)
);
