CREATE TABLE IF NOT EXISTS user_kyc_details
(
  id            UUID PRIMARY KEY   DEFAULT gen_random_uuid(),
  revision      INT NOT NULL,
  user_id       UUID NOT NULL,
  country_code  TEXT NOT NULL,
  first_name     TEXT NOT NULL,
  last_name     TEXT NOT NULL,
  gender        INT NOT NULL,
  date_of_birth DATE,
  address       JSONB,
  created_at    TIMESTAMP NOT NULL DEFAULT now(),
  INDEX user_id_ind (user_id),
  CONSTRAINT  user_id_revision_uniq UNIQUE (user_id, revision)
)