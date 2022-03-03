CREATE TABLE IF NOT EXISTS identities (
    id                          UUID PRIMARY KEY,
    email                       TEXT NOT NULL,
    first_name                  TEXT NOT NULL,
    last_name                   TEXT NOT NULL,
    country_id                  UUID NOT NULL,
    mobile_number               TEXT NOT NULL,
    created_at                  TIMESTAMP NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMP NOT NULL DEFAULT now()
);
