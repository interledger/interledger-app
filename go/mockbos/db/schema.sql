CREATE TABLE public.xago_sub_accounts
(
    id                           UUID      DEFAULT gen_random_uuid()                    NOT NULL,
    deposit_address              TEXT      DEFAULT 'rU8cx5WJMFp2n4tQ3Engzzr8Nnqm3EsSfC' NOT NULL,
    deposit_tag                  TEXT      DEFAULT 'tag'                                NOT NULL,
    first_name                   VARCHAR(255)                                           NOT NULL,
    last_name                    VARCHAR(255)                                           NOT NULL,
    email                        VARCHAR(255)                                           NOT NULL,
    mobile_number                VARCHAR(15) UNIQUE                                     NOT NULL,
    country                      CHAR(2)                                                NOT NULL,
    nationality                  CHAR(2)                                                NOT NULL,
    identification_document_type VARCHAR(255)                                           NOT NULL,
    identification_number        VARCHAR(255)                                           NOT NULL,
    address                      VARCHAR(255)                                           NOT NULL,
    city                         VARCHAR(255)                                           NOT NULL,
    district                     VARCHAR(255)                                           NOT NULL,
    postal_code                  CHAR(5)                                                NOT NULL,
    address_document_type        VARCHAR(255)                                           NOT NULL,
    date_of_birth                DATE                                                   NOT NULL,
    created_at                   TIMESTAMP DEFAULT now()::TIMESTAMP                     NOT NULL,
    updated_at                   TIMESTAMP DEFAULT now()::TIMESTAMP                     NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE public."xago_beneficiaries"
(
    id                           UUID      default gen_random_uuid() NOT NULL,
    name                         TEXT,
    scope                        TEXT,
    currency_code                TEXT,
    account_number               TEXT,
    branch_code                  TEXT,
    bank_name                    TEXT,
    bank_country                 TEXT,
    account_name                 TEXT,
    bank_beneficiary_type        TEXT,
    reference                    TEXT,
    iban                         TEXT,
    bic                          TEXT,
    beneficiary_physical_address TEXT,
    beneficiary_district         TEXT,
    beneficiary_city             TEXT,
    beneficiary_country          TEXT,
    beneficiary_postal_code      TEXT,
    beneficiary_address          TEXT,
    account_type                 TEXT,
    created_at                   TIMESTAMP DEFAULT now()::TIMESTAMP  NOT NULL,
    updated_at                   TIMESTAMP DEFAULT now()::TIMESTAMP  NOT NULL,
    primary key (id)
);

CREATE TABLE public."xago_transactions"
(
    id              UUID      default gen_random_uuid() NOT NULL,
    currency_code   text                                not null,
    amount          float                               not null,
    origin_amount   float                               not null,
    status          text                                not null,
    beneficiary_id  text                                not null,
    idempotency_key text                         not null,
    type            text                                not null,
    created_at      TIMESTAMP DEFAULT now()::TIMESTAMP  NOT NULL,
    updated_at      TIMESTAMP DEFAULT now()::TIMESTAMP  NOT NULL,
    primary key (id)
);

CREATE TABLE public.pti_users
(
    id UUID default gen_random_uuid() NOT NULL,
    type TEXT,
    date_of_birth TEXT,
    source_of_funds TEXT,
    first_name TEXT,
    middle_name TEXT,
    last_name TEXT,
    state TEXT,
    created_at      TIMESTAMP DEFAULT now()::TIMESTAMP  NOT NULL,
    updated_at      TIMESTAMP DEFAULT now()::TIMESTAMP  NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE public.pti_user_emails
(
    id UUID default gen_random_uuid() NOT NULL,
    user_id UUID NOT NULL,
    address TEXT NOT NULL,
    is_default BOOLEAN,
    created_at      TIMESTAMP DEFAULT now()::TIMESTAMP  NOT NULL,
    updated_at      TIMESTAMP DEFAULT now()::TIMESTAMP  NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE public.pti_user_addresses
(
    id UUID default gen_random_uuid() NOT NULL,
    user_id UUID NOT NULL,
    street TEXT,
    postal_code TEXT,
    state_code TEXT,
    country TEXT,
    is_default BOOLEAN,
    created_at      TIMESTAMP DEFAULT now()::TIMESTAMP  NOT NULL,
    updated_at      TIMESTAMP DEFAULT now()::TIMESTAMP  NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE public.pti_user_phones
(
    id UUID default gen_random_uuid() NOT NULL,
    user_id UUID NOT NULL,
    number TEXT,
    is_default BOOLEAN,
    created_at      TIMESTAMP DEFAULT now()::TIMESTAMP  NOT NULL,
    updated_at      TIMESTAMP DEFAULT now()::TIMESTAMP  NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE public.pti_wallets
(
    id TEXT NOT NULL,
    user_id UUID NOT NULL,
    balance BIGINT default 0,
    currency TEXT,
    created_at      TIMESTAMP DEFAULT now()::TIMESTAMP  NOT NULL,
    updated_at      TIMESTAMP DEFAULT now()::TIMESTAMP  NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE public.pti_transactions
(
    id UUID default gen_random_uuid() NOT NULL,
    request_id TEXT NOT NULL,
    user_id UUID NOT NULL,
    client_id TEXT NOT NULL,
    date TEXT,
    status TEXT,
    transaction_type TEXT,
    payment_method TEXT,
    payment_status_provider_response_code TEXT,
    payment_status_provider_response_category TEXT,
    amount BIGINT,
    currency TEXT,
    total_amount BIGINT,
    total_currency TEXT,
    fee_amount BIGINT,
    fee_currency TEXT,
    subtotal_amount BIGINT,
    subtotal_currency TEXT,
    created_at      TIMESTAMP DEFAULT now()::TIMESTAMP  NOT NULL,
    updated_at      TIMESTAMP DEFAULT now()::TIMESTAMP  NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (request_id)
);

CREATE TABLE public.atlas_schema_history
(
    id         UUID      DEFAULT gen_random_uuid() NOT NULL,
    hash       TEXT                                NOT NULL,
    created_at TIMESTAMP DEFAULT now()::TIMESTAMP  NOT NULL
);
