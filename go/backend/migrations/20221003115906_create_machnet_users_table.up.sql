CREATE TABLE IF NOT EXISTS machnet_users (
	id 				UUID PRIMARY KEY,
	wallet_id 		UUID NOT NULL REFERENCES wallets,
	created_at   	TIMESTAMP NOT NULL DEFAULT now(),
	updated_at   	TIMESTAMP NOT NULL DEFAULT now()
);
