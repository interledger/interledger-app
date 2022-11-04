ALTER TABLE openpayments_incoming_payment DROP COLUMN IF EXISTS from_payment_pointer_id;
ALTER TABLE openpayments_outgoing_payment DROP COLUMN IF EXISTS to_payment_pointer_id;
ALTER TABLE openpayments_incoming_payment DROP COLUMN IF EXISTS description;