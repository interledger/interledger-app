CREATE schema IF NOT EXISTS public;

CREATE TABLE public.ledgers
(
    id         BIGINT    NOT NULL,
    name       TEXT,
    asset      TEXT      NOT NULL,
    scale      BIGINT    NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id)
);

CREATE TABLE public.ledger_accounts
(
    id                             UUID      NOT NULL DEFAULT gen_random_uuid(),
    ledger_id                      BIGINT    NOT NULL,
    code                           BIGINT    NOT NULL,
    debits_must_not_exceed_credits BOOLEAN   NOT NULL DEFAULT false,
    credits_must_not_exceed_debits BOOLEAN   NOT NULL DEFAULT false,
    debits_pending                 BIGINT    NOT NULL DEFAULT 0,
    debits_posted                  BIGINT    NOT NULL DEFAULT 0,
    credits_pending                BIGINT    NOT NULL DEFAULT 0,
    credits_posted                 BIGINT    NOT NULL DEFAULT 0,
    created_at                     TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at                     TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id),
    CONSTRAINT fk_accounts_ledgers FOREIGN KEY (ledger_id) REFERENCES public.ledgers (id) ON UPDATE NO ACTION ON DELETE NO ACTION
);

CREATE TABLE public.ledger_transfers
(
    id                UUID      NOT NULL DEFAULT gen_random_uuid(),
    ledger_id         BIGINT    NOT NULL,
    code              BIGINT    NOT NULL,
    debit_account_id  UUID      NOT NULL,
    credit_account_id UUID      NOT NULL,
    amount            BIGINT    NOT NULL,
    state             SMALLINT  NOT NULL,
    timeout_at        TIMESTAMP,
    created_at        TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id),
    CONSTRAINT fk_transfers_accounts_credits FOREIGN KEY (credit_account_id) REFERENCES public.ledger_accounts (id) ON UPDATE NO ACTION ON DELETE NO ACTION,
    CONSTRAINT fk_transfers_accounts_debits FOREIGN KEY (debit_account_id) REFERENCES public.ledger_accounts (id) ON UPDATE NO ACTION ON DELETE NO ACTION
);
