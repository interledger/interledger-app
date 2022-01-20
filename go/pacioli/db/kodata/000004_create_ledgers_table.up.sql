CREATE TABLE IF NOT EXISTS ledgers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT,
    created_at  TIMESTAMP NOT NULL DEFAULT now(),
    updated_at  TIMESTAMP NOT NULL DEFAULT now()
);
