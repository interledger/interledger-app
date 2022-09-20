CREATE TABLE IF NOT EXISTS wallets (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT NOT NULL,
  created_at  TIMESTAMP NOT NULL DEFAULT now(),
  updated_at  TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_wallets (
  user_id UUID NOT NULL,
  wallet_id UUID NOT NULL,
  PRIMARY KEY (user_id, wallet_id),
  CONSTRAINT fk_user_wallets_wallets
    FOREIGN KEY(wallet_id) REFERENCES wallets(id)
);