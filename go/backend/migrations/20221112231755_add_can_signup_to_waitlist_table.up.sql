ALTER TABLE waitlist_signups ADD COLUMN IF NOT EXISTS can_signup BOOLEAN NOT NULL default false;
ALTER TABLE waitlist_signups ADD COLUMN IF NOT EXISTS user_id UUID;