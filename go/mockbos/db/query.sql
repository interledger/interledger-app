-- name: CreateXagoSubAccount :one
INSERT INTO xago_sub_accounts
(id, deposit_tag, first_name, last_name, email, mobile_number,
 country, nationality, identification_document_type, identification_number,
 address, city, district, postal_code, address_document_type, date_of_birth, deposit_reference)
VALUES (gen_random_uuid(), $1, $2, $3, $4,
        $5, $6, $7, $8,
        $9, $10, $11, $12, $13, $14, $15, $16)
returning id;

-- name: GetXagoSubAccount :one
select * from xago_sub_accounts where id = $1 limit 1;

-- name: GetXagoSubAccountByDepositReference :one
select * from xago_sub_accounts where deposit_reference = $1 limit 1;

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
