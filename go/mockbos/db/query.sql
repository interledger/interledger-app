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
values (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7)
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
