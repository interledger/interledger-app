CREATE TABLE IF NOT EXISTS signups
(
	id										UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	first_name						TEXT,
	last_name							TEXT,
	country_code        	VARCHAR(2),
	email									TEXT,
	mobile_number					TEXT,
	user_id               UUID,
	created_at						TIMESTAMP NOT NULL DEFAULT now(),
	updated_at						TIMESTAMP NOT NULL DEFAULT now()
);
