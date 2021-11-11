CREATE TABLE IF NOT EXISTS organisations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT,
    owner_id    UUID NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT now(),
    updated_at  TIMESTAMP NOT NULL DEFAULT now()
);
