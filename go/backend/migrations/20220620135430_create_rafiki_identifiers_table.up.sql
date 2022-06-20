CREATE TABLE IF NOT EXISTS rafiki_identifiers
(
	id			TEXT PRIMARY KEY,
	account_id	UUID NOT NULL,
	asset_code	TEXT NOT NULL,
	asset_scale	INT8 NOT NULL,
	created_at	TIMESTAMP NOT NULL DEFAULT now(),
	updated_at	TIMESTAMP NOT NULL DEFAULT now()
);
