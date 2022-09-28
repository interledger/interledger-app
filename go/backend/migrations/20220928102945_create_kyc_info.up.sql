CREATE TABLE IF NOT EXISTS user_kyc_details
(
  id            UUID PRIMARY KEY   DEFAULT gen_random_uuid(),
  revision      INT NOT NULL,
  user_id       UUID NOT NULL,
  country_code  TEXT,
  first_name    TEXT,
  last_name     TEXT,
  gender        INT,
  date_of_birth DATE,
  address       JSONB,
  created_at    TIMESTAMP NOT NULL DEFAULT now(),
  INDEX user_id_ind (user_id),
  CONSTRAINT  user_id_revision_uniq UNIQUE (user_id, revision)
)