ALTER TABLE openpayments_outgoing_payment DROP COLUMN IF EXISTS completed;
DROP TABLE IF EXISTS openpayments_transactions;