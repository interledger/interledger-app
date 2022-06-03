CREATE TABLE IF NOT EXISTS unit_counterparties
(
	id 						UUID PRIMARY KEY,
	unit_counterparty_id	TEXT NOT NULL,
	created_at				TIMESTAMP NOT NULL DEFAULT now(),
	updated_at				TIMESTAMP NOT NULL DEFAULT now()
);
