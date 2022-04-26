CREATE TABLE IF NOT EXISTS withdrawals
(
    id                UUID PRIMARY KEY   DEFAULT gen_random_uuid(),
    account_id        UUID      NOT NULL,
    funding_source_id UUID      NOT NULL,
    amount            BIGINT    NOT NULL,
    state             TEXT               NOT NULL,
    created_at        TIMESTAMP NOT NULL DEFAULT now(),
    updated_at        TIMESTAMP NOT NULL DEFAULT now()
);