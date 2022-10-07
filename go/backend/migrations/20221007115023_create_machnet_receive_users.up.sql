CREATE TABLE IF NOT EXISTS machnet_receive_users (
	id					UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	receive_wallet_id	UUID NOT NULL REFERENCES wallets(id),
	send_user_id		UUID NOT NULL REFERENCES machnet_users(id),
    created_at			TIMESTAMP NOT NULL DEFAULT now(),
    updated_at			TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT  send_user_id_receive_wallet_id_uniq UNIQUE (receive_wallet_id, send_user_id),
    INDEX receive_wallet_id_ind (receive_wallet_id)
);
