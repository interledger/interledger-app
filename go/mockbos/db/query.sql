-- name: CreateXagoSubAccount :one
INSERT INTO xago_sub_accounts
(id, deposit_tag, first_name, last_name, email, mobile_number,
 country, nationality, identification_document_type, identification_number,
 address, city, district, postal_code, address_document_type, date_of_birth)
VALUES (gen_random_uuid(), $1, $2, $3, $4,
        $5, $6, $7, $8,
        $9, $10, $11, $12, $13, $14, $15)
returning id;

-- name: GetXagoSubAccount :one
select * from xago_sub_accounts where id = $1 limit 1;

-- name: GetXagoSubAccountByDepositReference :one
select * from xago_sub_accounts where deposit_tag = $1 limit 1;

-- name: CreateXagoBeneficiary :one
INSERT INTO xago_beneficiaries
(id, name, scope, currency_code, account_number, branch_code,
 bank_name, bank_country, account_name, bank_beneficiary_type,
 reference, iban, bic, beneficiary_physical_address,
 beneficiary_district, beneficiary_city, beneficiary_country, beneficiary_postal_code,
 beneficiary_address, account_type)
VALUES (gen_random_uuid(),
        $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
returning id;

-- name: ListXagoBeneficiaries :many
select *
from xago_beneficiaries;

-- name: GetXagoBeneficiary :one
select *
from xago_beneficiaries
where id = $1;

-- name: CreateXagoTransaction :one
insert into xago_transactions
(id, currency_code, amount, origin_amount, status, beneficiary_id, idempotency_key, type)
values ($1, $2, $3, $4, $5, $6, $7, $8)
returning *;

-- name: GetXagoTransaction :one
select *
from xago_transactions
where id = $1;

-- name: ListXagoDeposits :many
select *
from xago_transactions
where type = 'deposit'
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CreatePTIUser :one
INSERT INTO pti_users
(id, type, first_name, middle_name, last_name, date_of_birth, source_of_funds, state) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *;

-- name: CreatePTIUserEmail :copyfrom
INSERT INTO pti_user_emails (user_id, address, is_default) VALUES ($1, $2, $3);

-- name: ListPTIUserEmails :many
SELECT * FROM pti_user_emails WHERE id = $1;

-- name: ListPTIUserAddresses :many
SELECT * FROM pti_user_addresses WHERE id = $1;

-- name: CreatePTIUserAddress :copyfrom
INSERT INTO pti_user_addresses (user_id, street, postal_code, state_code, country, is_default) VALUES ($1, $2, $3, $4, $5, $6);

-- name: CreatePTIUserPhone :copyfrom
INSERT INTO pti_User_phones (user_id, number, is_default) VALUES ($1, $2, $3);

-- name: CreatePTIWallet :one
INSERT INTO pti_wallets (id, user_id, currency) VALUES ($1, $2, $3) RETURNING *;

-- name: GetPTIUserWallets :many
SELECT * FROM pti_wallets WHERE user_id=$1;

-- name: GetPTIUser :one
SELECT * FROM pti_users WHERE id=$1;

-- name: GetPTIUserEmails :many
SELECT * FROM pti_user_emails WHERE user_id=$1;

-- name: GetPTIUserAddresses :many
SELECT * FROM pti_user_addresses WHERE user_id=$1;

-- name: GetPTIUserWallet :one
SELECT * FROM pti_wallets WHERE id=$1 AND user_id=$2;

-- name: GetPTIWallet :one
SELECT * FROM pti_wallets WHERE id=$1;

-- name: GetPTIUserPhones :many
SELECT * FROM pti_user_phones WHERE user_id=$1;

-- name: CreatePTITransaction :one
INSERT INTO pti_transactions (request_id, user_id, client_id, date, status, transaction_type, payment_method, payment_status_provider_response_code, payment_status_provider_response_category, amount, currency, total_amount, total_currency, fee_amount, fee_currency, subtotal_amount, subtotal_currency) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17) RETURNING *;

-- name: AdjustPTIWalletBalance :exec
UPDATE pti_wallets SET balance=balance+$1 WHERE id=$2;

-- name: GetPTITransactionByRequestID :one
SELECT * FROM pti_transactions WHERE request_id=$1;

-- name: CreateAstraIntent :one
INSERT INTO astra_user_intents (
    email,
    phone,
    first_name,
    last_name,
    address1,
    address2,
    city,
    state,
    postal_code,
    date_of_birth,
    ssn,
    ip_address
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: GetAstraIntent :one
SELECT * FROM astra_user_intents WHERE id=$1;

-- name: CreateAstraUser :one
INSERT INTO astra_users (
    email,
    phone,
    first_name,
    last_name,
    address1,
    address2,
    city,
    state,
    postal_code,
    date_of_birth,
    ssn,
    ip_address,
    status,
    kyc_type
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
) RETURNING *;

-- name: UpdateAstraUserIntentUserID :exec
UPDATE astra_user_intents SET user_id=$1 WHERE id=$2;

-- name: GetAstraUser :one
SELECT * FROM astra_users WHERE id=$1;

-- name: CreateAstraAccessToken :one
INSERT INTO astra_access_tokens (user_id, expires_in, token_type) VALUES ($1, $2, $3) RETURNING *;

-- name: GetAstraAccessToken :one
SELECT * FROM astra_access_tokens WHERE id=$1;

-- name: GetAstraAccessTokenByRefreshToken :one
SELECT * FROM astra_access_tokens WHERE refresh_token=$1;

-- name: CreateAstraUserCard :one
INSERT INTO astra_user_cards (
        user_id,
        address_verified,
        card_company,
        city,
        expiration_date,
        first_name,
        first_six_digits,
        last_four_digits,
        last_name,
        card_type,
        pull_enabled,
        push_enabled,
        removed,
        review_status,
        state,
        status,
        street_line_1,
        street_line_2,
        zip_code
    ) VALUES (
        $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
    ) RETURNING *;

-- name: GetAstraUserCard :one
SELECT * FROM astra_user_cards WHERE id=$1 AND user_id=$2;

-- name: GetAstraCard :one
SELECT * FROM astra_user_cards WHERE id=$1;

-- name: CreateAstraUserAccount :one
INSERT INTO astra_user_accounts (
    user_id,
    official_name,
    name,
    mask,
    institution_name,
    institution_logo,
    type,
    subtype,
    connection_status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetAstraUserAccount :one
SELECT * FROM astra_user_accounts WHERE id=$1 AND user_id=$2;

-- name: GetAstraAccount :one
SELECT * FROM astra_user_accounts WHERE id=$1;

-- name: CreateAstraTransaction :one
INSERT INTO astra_transactions (
    id,
    routine_type,
    routine_name,
    routine_id,
    client_correlation_id,
    source_id,
    destination_id,
    destination_user_id,
    amount,
    payment_type,
    initiated,
    updated,
    estimated_clearing_date,
    astra_settlement_reason,
    failure_reason,
    chargeback_action_status,
    chargeback_cb_id,
    chargeback_created,
    chargeback_exception_code,
    chargeback_exception_date,
    chargeback_exception_description,
    chargeback_exception_id,
    chargeback_exception_settled_amount,
    chargeback_exception_type,
    chargeback_merchant_reference_id,
    chargeback_network_id,
    chargeback_original_creation_date,
    chargeback_original_processed_date,
    chargeback_original_settled_amount,
    chargeback_status_date,
    chargeback_updated,
    chargeback_user_id,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
    $21, $22, $23, $24, $25, $26, $27, $28, $29, $30,
    $31, $32, $33
) RETURNING *;

-- name: GetAstraTransaction :one
SELECT * FROM astra_transactions WHERE id=$1;
