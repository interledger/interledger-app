CREATE TABLE IF NOT EXISTS funding_sources (
	id 							UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	wallet_id                   UUID NOT NULL,
	name 						TEXT NOT NULL,
	mask						TEXT NOT NULL,
	provider			            TEXT NOT NULL,
	type                     TEXT NOT NULL,
	created_at                  TIMESTAMP NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMP NOT NULL DEFAULT now()
);
