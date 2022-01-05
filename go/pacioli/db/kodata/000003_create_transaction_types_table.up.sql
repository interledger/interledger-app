CREATE TABLE IF NOT EXISTS transaction_types(
	id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    name        TEXT NOT NULL,
    description TEXT,
    credit_account_category_code INT NOT NULL,
    debit_account_category_code INT NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT now(),
    updated_at  TIMESTAMP NOT NULL DEFAULT now()
);