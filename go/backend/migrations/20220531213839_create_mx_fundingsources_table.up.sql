CREATE TABLE IF NOT EXISTS mx_fundingsources
(
	id				UUID PRIMARY KEY,
	account_id		UUID NOT NULL,
	mx_user_guid	TEXT NOT NULL,
	mx_member_guid	TEXT NOT NULL,
	mx_account_guid	TEXT NOT NULL
);
