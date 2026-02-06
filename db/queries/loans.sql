-- name: GetLoanByID :one
SELECT *
FROM loans
WHERE id = $1;
-- name: InsertLoan :one
INSERT INTO loans (
    principal_amount,
    total_interest_amount,
    total_payable_amount,
    weekly_payment_amount,
    total_weeks,
    start_date
  )
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;