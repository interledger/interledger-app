CREATE TABLE IF NOT EXISTS unit_user_ach_deposits (
	id 					TEXT NOT NULL,
	deposit_account_id	TEXT NOT NULL,
	deposit_id			UUID NOT NULL,
	counterparty_id		UUID NOT NULL,
	amount				BIGINT NOT NULL,
	created_at			TIMESTAMP NOT NULL DEFAULT now(),
	updated_at			TIMESTAMP NOT NULL DEFAULT now()
);
