CREATE TABLE IF NOT EXISTS mx_accounts
(
	guid				TEXT NOT NULL,
	user_guid			TEXT NOT NULL,
	member_guid			TEXT NOT NULL,
	account_id			UUID NOT NULL,
	fundingsource_id	UUID NOT NULL,
	created_at			TIMESTAMP NOT NULL DEFAULT now(),
	updated_at			TIMESTAMP NOT NULL DEFAULT now(),
	PRIMARY KEY (guid, account_id)
);
