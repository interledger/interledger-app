CREATE TABLE IF NOT EXISTS ledger_accounts (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  ledger_id       INT NOT NULL,
  code            INT NOT NULL,
  flags           SMALLINT NULL,
  debits_pending  BIGINT NOT NULL DEFAULT 0,
  debits_posted   BIGINT NOT NULL DEFAULT 0,
  credits_pending BIGINT NOT NULL DEFAULT 0,
  credits_posted  BIGINT NOT NULL DEFAULT 0,
  created_at      TIMESTAMP NOT NULL DEFAULT now(),
  updated_at      TIMESTAMP NOT NULL DEFAULT now()
);


CREATE TABLE IF NOT EXISTS ledger_transfers (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  ledger_id         UUID NOT NULL,
  debit_account_id  UUID NOT NULL,
  credit_account_id UUID NOT NULL,
  pending_id        UUID NULL,
  amount            BIGINT NOT NULL, -- Check on Cockroach types
  state             SMALLINT NOT NULL, -- See code for state definitions
  timeout_at        TIMESTAMP NULL,
  created_at        TIMESTAMP NOT NULL DEFAULT now(),
  updated_at        TIMESTAMP NOT NULL DEFAULT now()
);