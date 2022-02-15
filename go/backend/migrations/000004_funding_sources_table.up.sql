CREATE TABLE IF NOT EXISTS funding_sources (
	id 							UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	identity_id                 UUID NOT NULL,
	name 						TEXT NOT NULL,
	mask						TEXT NOT NULL,
	verification_state          TEXT NOT NULL,
	type			            TEXT NOT NULL,
	type_id          			UUID NOT NULL,
	subtype                     TEXT NOT NULL,
	created_at                  TIMESTAMP NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMP NOT NULL DEFAULT now()
);