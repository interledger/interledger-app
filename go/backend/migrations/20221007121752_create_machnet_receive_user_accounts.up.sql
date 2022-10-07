CREATE TABLE IF NOT EXISTS machnet_receive_user_accounts (
	id					UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	receive_account_id	UUID NOT NULL REFERENCES machnet_receive_accounts(id),
	receive_user_id		UUID NOT NULL REFERENCES machnet_receive_users(id),
    created_at			TIMESTAMP NOT NULL DEFAULT now(),
    updated_at			TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT  receive_user_id_receive_account_id_uniq UNIQUE (receive_account_id, receive_user_id),
    INDEX receive_account_id_ind (receive_account_id)
);
