CREATE TABLE IF NOT EXISTS unit_customers
(
	id			TEXT PRIMARY KEY NOT NULL,
	account_id	UUID NOT NULL,
	type		TEXT NOT NULL,
	created_at 			TIMESTAMP NOT NULL DEFAULT now(),
	updated_at 			TIMESTAMP NOT NULL DEFAULT now()
);
