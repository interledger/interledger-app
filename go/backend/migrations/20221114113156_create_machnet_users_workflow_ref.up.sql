CREATE TABLE IF NOT EXISTS machnet_users_workflow_ref(
  id				      UUID PRIMARY KEY default gen_random_uuid(),
  user_id	        UUID NOT NULL REFERENCES machnet_users(id),
  workflow_id		  TEXT NOT NULL,
  workflow_run_id	TEXT NOT NULL,
  activity_name   TEXT NOT NULL,
  completed       BOOLEAN NOT NULL DEFAULT false,
  created_at			TIMESTAMP NOT NULL DEFAULT now(),
  updated_at			TIMESTAMP NOT NULL DEFAULT now(),
  INDEX user_id_ind(user_id)
);
