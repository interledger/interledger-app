-- name: GetPayment :one
SELECT * FROM payments
WHERE id = $1 LIMIT 1;