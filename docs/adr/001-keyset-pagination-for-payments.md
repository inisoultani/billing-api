# ADR-001: Keyset pagination for payments

## Context

Payments are queried per loan and paginated by paid_at + id.
OFFSET-based pagination was rejected due to performance degradation
on large datasets.

## Decision

We use keyset pagination with a composite index:

CREATE INDEX idx_repayments_loan_paidat_id
ON payments (loan_id, paid_at, id);

Query pattern:

- loan_id = ?
- (paid_at, id) > (?, ?)
- ORDER BY paid_at, id

## Evidence

EXPLAIN ANALYZE shows Bitmap Index Scan on idx_repayments_loan_paidat_id
with sub-millisecond execution time.

```Bash
QUERY PLAN                                                                                                                                      |
------------------------------------------------------------------------------------------------------------------------------------------------+
Limit  (cost=9.53..9.54 rows=2 width=36) (actual time=0.077..0.081 rows=2 loops=1)                                                              |
  ->  Sort  (cost=9.53..9.54 rows=2 width=36) (actual time=0.068..0.068 rows=2 loops=1)                                                         |
        Sort Key: paid_at, id                                                                                                                   |
        Sort Method: quicksort  Memory: 25kB                                                                                                    |
        ->  Bitmap Heap Scan on payments  (cost=4.17..9.52 rows=2 width=36) (actual time=0.047..0.048 rows=2 loops=1)                           |
              Recheck Cond: ((loan_id = 1) AND (ROW(paid_at, id) > ROW('2026-02-05 17:50:06.615139'::timestamp without time zone, 2)))          |
              Heap Blocks: exact=1                                                                                                              |
              ->  Bitmap Index Scan on idx_repayments_loan_paidat_id  (cost=0.00..4.17 rows=2 width=0) (actual time=0.032..0.032 rows=2 loops=1)|
                    Index Cond: ((loan_id = 1) AND (ROW(paid_at, id) > ROW('2026-02-05 17:50:06.615139'::timestamp without time zone, 2)))      |
Planning Time: 0.251 ms                                                                                                                         |
Execution Time: 0.140 ms                                                                                                                        |
```

## Consequences

- Stable pagination
- Predictable performance
- Requires cursor parameters
