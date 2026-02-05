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