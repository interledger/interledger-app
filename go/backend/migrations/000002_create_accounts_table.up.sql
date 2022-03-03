CREATE TABLE IF NOT EXISTS accounts (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    identity_id         UUID NOT NULL,
    ledger_account_id   UUID NOT NULL,
    provider            TEXT DEFAULT '',
    provider_id         TEXT DEFAULT '',
    verification_state  TEXT DEFAULT '',
    created_at          TIMESTAMP NOT NULL DEFAULT now(),
    updated_at          TIMESTAMP NOT NULL DEFAULT now()
);
