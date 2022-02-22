CREATE TABLE IF NOT EXISTS account_transactions (
	id							UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	account_id					UUID NOT NULL,
	description					TEXT DEFAULT '',
	type						TEXT NOT NULL,
	net_amount					INT NOT NULL,
	state						TEXT DEFAULT '',
	transfer_ids				STRING[] DEFAULT ARRAY[],
	created_at					TIMESTAMP NOT NULL DEFAULT now(),
    updated_at					TIMESTAMP NOT NULL DEFAULT now()
);