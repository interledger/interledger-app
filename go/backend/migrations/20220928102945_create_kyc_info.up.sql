CREATE TABLE IF NOT EXISTS individual_kyc_details
(
  id            UUID PRIMARY KEY   DEFAULT gen_random_uuid(),
  revision      INT NOT NULL,
  wallet_id     UUID NOT NULL,
  country_code  TEXT NOT NULL,
  first_name     TEXT NOT NULL,
  last_name     TEXT NOT NULL,
  gender        INT NOT NULL,
  date_of_birth DATE,
  address       JSONB,
  created_at    TIMESTAMP NOT NULL DEFAULT now(),
  INDEX wallet_id_ind (wallet_id),
  CONSTRAINT  wallet_id_revision_uniq UNIQUE (wallet_id, revision),
  CONSTRAINT fk_individual_kyc_details_wallets
    FOREIGN KEY(wallet_id) REFERENCES wallets(id)
)