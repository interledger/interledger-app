CREATE TABLE IF NOT EXISTS tenants (
	id          TEXT PRIMARY KEY,
	created_at  TIMESTAMP NOT NULL DEFAULT now(),
    updated_at  TIMESTAMP NOT NULL DEFAULT now()
);
