ALTER TABLE unit_user_ach_deposits DROP COLUMN counterparty_id;
ALTER TABLE unit_user_ach_deposits ADD counterparty_id TEXT NOT NULL;
