CREATE TABLE IF NOT EXISTS ledger_accounts (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  ledger_id       UUID NOT NULL,
  code            INT NOT NULL,
  debits_pending  TEXT,
  debits_posted   TEXT,
  credits_pending TEXT,
  credits_posted  TEXT,
  created_at      TIMESTAMP NOT NULL DEFAULT now(),
  updated_at      TIMESTAMP NOT NULL DEFAULT now()
);


CREATE TABLE IF NOT EXISTS ledger_transfers (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  ledger_id         UUID NOT NULL,
  debit_account_id  UUID NOT NULL,
  credit_account_id UUID NOT NULL,
  amount            TEXT NOT NULL, -- Check on Cockroach types
  state             INT NOT NULL, -- See code for state definitions
  created_at        TIMESTAMP NOT NULL DEFAULT now(),
  updated_at        TIMESTAMP NOT NULL DEFAULT now()
);