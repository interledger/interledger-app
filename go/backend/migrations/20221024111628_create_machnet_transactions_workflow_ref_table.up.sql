CREATE TABLE IF NOT EXISTS machnet_transactions_workflow_ref(
	id				UUID PRIMARY KEY,
	send_user_id	UUID NOT NULL REFERENCES machnet_users(id),
	workflow_id		TEXT NOT NULL,
	workflow_run_id	TEXT NOT NULL,
	created_at			TIMESTAMP NOT NULL DEFAULT now(),
    updated_at			TIMESTAMP NOT NULL DEFAULT now()
);
