CREATE TABLE IF NOT EXISTS machnet_receive_user_bank_accounts (
	id                         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	receive_bank_account_id	   UUID NOT NULL REFERENCES machnet_receive_bank_accounts(id),
	receive_user_id            UUID NOT NULL REFERENCES machnet_receive_users(id),
    created_at                 TIMESTAMP NOT NULL DEFAULT now(),
    updated_at                 TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT  receive_user_id_receive_bank_account_id_uniq UNIQUE (receive_bank_account_id, receive_user_id),
    INDEX receive_bank_account_id_ind (receive_bank_account_id)
);
