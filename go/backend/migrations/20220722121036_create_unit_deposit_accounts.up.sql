CREATE TABLE IF NOT EXISTS unit_deposit_accounts
(
	id				TEXT PRIMARY KEY,
    customer_id 	TEXT NOT NULL,
    created_at		TIMESTAMP NOT NULL DEFAULT now(),
    updated_at		TIMESTAMP NOT NULL DEFAULT now()
);
