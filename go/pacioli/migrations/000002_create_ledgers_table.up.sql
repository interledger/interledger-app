CREATE TABLE IF NOT EXISTS ledgers (
    id          INT PRIMARY KEY,
    name        TEXT,
    asset       TEXT NOT NULL,
    scale       INT NOT NULL, 
    created_at  TIMESTAMP NOT NULL DEFAULT now(),
    updated_at  TIMESTAMP NOT NULL DEFAULT now()
);
