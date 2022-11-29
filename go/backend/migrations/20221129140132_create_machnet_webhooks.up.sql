CREATE TABLE IF NOT EXISTS machnet_webhook(
  id				      UUID PRIMARY KEY default gen_random_uuid(),
  user_id	        UUID NOT NULL REFERENCES machnet_users(id),
  event_name		  TEXT NOT NULL,
  resource_id	    TEXT NOT NULL,
  subscription_id TEXT NOT NULL,
  payload         JSON NULL,
  created_at			TIMESTAMP NOT NULL DEFAULT now()
);
