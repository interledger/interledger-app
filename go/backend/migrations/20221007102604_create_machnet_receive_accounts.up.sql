CREATE TABLE IF NOT EXISTS machnet_receive_accounts (
	id				UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	wallet_id		UUID NOT NULL REFERENCES wallets(id),
	account_number	TEXT NOT NULL,
	type			TEXT NOT NULL,
	bank_id			INT NOT NULL,
	branch_id		INT NOT NULL,
    created_at		TIMESTAMP NOT NULL DEFAULT now(),
    updated_at		TIMESTAMP NOT NULL DEFAULT now()
);
