ALTER TABLE waitlist_signups ADD COLUMN mug_id TEXT;
ALTER TABLE waitlist_signups ADD CONSTRAINT  waitlist_signups_mug_id UNIQUE (mug_id);