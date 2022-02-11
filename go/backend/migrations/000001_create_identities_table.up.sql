CREATE TABLE IF NOT EXISTS identities (
    id                          UUID PRIMARY KEY,
    email                       TEXT NOT NULL,
    first_name                  TEXT NOT NULL,
    last_name                   TEXT NOT NULL,
    country_id                  UUID NOT NULL,
    mobile_number               TEXT NOT NULL,
    date_of_birth               DATE DEFAULT DATE '0000-01-01',
    address                     STRING[] DEFAULT ARRAY[],
    state                       TEXT DEFAULT '',
    city                        TEXT DEFAULT '',
    postal_code                 TEXT DEFAULT '',
    tax_id_number               TEXT DEFAULT '',
    provider                    TEXT DEFAULT '',
    provider_id                 TEXT DEFAULT '',
    verification_state          TEXT DEFAULT '',
    created_at                  TIMESTAMP NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMP NOT NULL DEFAULT now()
);
