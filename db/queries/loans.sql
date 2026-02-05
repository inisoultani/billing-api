-- name: GetLoanByID :one
SELECT *
FROM loans
WHERE id = $1;
-- name: GetTotalPaidAmount :one
SELECT COALESCE(SUM(amount), 0)::BIGINT AS total_paid
FROM payments
WHERE loan_id = $1;
-- name: InsertPayment :one
INSERT INTO payments (loan_id, week_number, amount, paid_at)
VALUES ($1, $2, $3, $4)
RETURNING *;
-- name: GetLastPaidWeek :one
SELECT COALESCE(MAX(week_number), 0)::BIGINT AS last_paid_week
FROM payments
WHERE loan_id = $1;
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
-- name: GetPaidWeeksCount :one
SELECT COUNT(*)::INT
FROM payments
WHERE loan_id = $1;