CREATE TABLE IF NOT EXISTS tenants (
	id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	identifier  TEXT NOT NULL UNIQUE
);