CREATE TABLE IF NOT EXISTS onboarding
(
	id										UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	first_name						TEXT DEFAULT '',
	last_name							TEXT DEFAULT '',
	country_of_residence	TEXT DEFAULT '',
	email									TEXT DEFAULT '',
	phone									TEXT DEFAULT '',
	phone_verified				BOOLEAN DEFAULT false,
	service_agreement			BOOLEAN DEFAULT false,
	created_at						TIMESTAMP NOT NULL DEFAULT now(),
	updated_at						TIMESTAMP NOT NULL DEFAULT now()
);
