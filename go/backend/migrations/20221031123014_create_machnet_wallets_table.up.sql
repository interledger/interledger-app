CREATE TABLE IF NOT EXISTS machnet_wallets (
	id					UUID PRIMARY KEY,
	send_user_id		UUID NOT NULL references machnet_users(id),
	nickname			TEXT NOT NULL,
	created_at			TIMESTAMP NOT NULL DEFAULT now(),
	updated_at			TIMESTAMP NOT NULL DEFAULT now(),
	CONSTRAINT  send_user_id_nickname_uniq UNIQUE (send_user_id, nickname)
);
