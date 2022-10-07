CREATE TABLE IF NOT EXISTS waitlist_signups
(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    country_code CHAR(2) NOT NULL,
    full_name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    notified_at TIMESTAMP,
    notified_count INT,
    can_signup BOOLEAN NOT NULL default false,
    user_id UUID,
    CONSTRAINT  email_country_code UNIQUE (email, country_code)
);