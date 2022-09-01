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
  updated_at      TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT fk_accounts_ledgers
    FOREIGN KEY(ledger_id) REFERENCES ledgers(id)
);


CREATE TABLE IF NOT EXISTS ledger_transfers (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  ledger_id         INT NOT NULL,
  code              INT NOT NULL,
  debit_account_id  UUID NOT NULL,
  credit_account_id UUID NOT NULL,
  pending_id        UUID NULL,
  amount            BIGINT NOT NULL,
  state             SMALLINT NOT NULL, -- See code for state definitions
  timeout_at        TIMESTAMP NULL,
  created_at        TIMESTAMP NOT NULL DEFAULT now(),
  updated_at        TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT fk_transfers_accounts_credits
    FOREIGN KEY(credit_account_id) REFERENCES ledger_accounts(id),
  CONSTRAINT fk_transfers_accounts_debits
    FOREIGN KEY(debit_account_id) REFERENCES ledger_accounts(id)
);