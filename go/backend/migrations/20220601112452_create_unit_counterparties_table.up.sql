CREATE TABLE IF NOT EXISTS unit_counterparties
(
	id 						TEXT NOT NULL,
	fundingsource_id		UUID NOT NULL,
	created_at				TIMESTAMP NOT NULL DEFAULT now(),
	updated_at				TIMESTAMP NOT NULL DEFAULT now(),
	PRIMARY KEY(id, fundingsource_id)
);
