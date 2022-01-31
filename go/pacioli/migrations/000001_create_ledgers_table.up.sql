CREATE TABLE IF NOT EXISTS ledgers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT,
    code        INT UNIQUE NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT now(),
    updated_at  TIMESTAMP NOT NULL DEFAULT now()
);
