CREATE TABLE IF NOT EXISTS unit_events
(
	id 							TEXT 			NOT NULL PRIMARY KEY,
	event_type 			TEXT 			NOT NULL,
	raw_event 			TEXT 			NOT NULL,
	created_at 			TIMESTAMP NOT NULL DEFAULT now(),
	updated_at 			TIMESTAMP NOT NULL DEFAULT now()
);