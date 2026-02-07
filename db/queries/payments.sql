-- name: ListPaymentsByLoanID :many
SELECT id,
  loan_id,
  week_number,
  amount,
  paid_at
FROM payments
WHERE loan_id = @loan_id::bigint
  AND (
    (
      sqlc.narg('cursor_paid_at')::timestamptz IS NULL
      AND sqlc.narg('cursor_id')::bigint IS NULL
    )
    OR (
      (paid_at, id) > (
        sqlc.narg('cursor_paid_at')::timestamptz,
        sqlc.narg('cursor_id')::bigint
      )
    )
  )
ORDER BY paid_at ASC,
  id ASC
LIMIT @limit_val::int;
-- name: GetTotalPaidAmount :one
SELECT COALESCE(SUM(amount), 0)::BIGINT AS total_paid
FROM payments
WHERE loan_id = $1;
-- name: InsertPayment :one
INSERT INTO payments (
    loan_id,
    week_number,
    amount,
    idempotency_key,
    paid_at
  )
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
-- name: GetLastPaidWeek :one
SELECT COALESCE(MAX(week_number), 0)::INT AS last_paid_week
FROM payments
WHERE loan_id = $1;
-- name: GetPaidWeeksCount :one
SELECT COUNT(*)::INT
FROM payments
WHERE loan_id = $1;