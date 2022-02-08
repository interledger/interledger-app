CREATE TABLE IF NOT EXISTS identities (
    id          UUID PRIMARY KEY,
    email       TEXT NOT NULL,
    legal_name  TEXT NOT NULL,
    country     TEXT NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT now(),
    updated_at  TIMESTAMP NOT NULL DEFAULT now()
);
