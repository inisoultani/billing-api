# ADR-002: Derived states are computed dynamically

## Status

Accepted

## Context

The system has domain states that can be _derived_ from existing data,
for example:

- delinquency status
- overdue indicators
- payment completeness

Persisting these states would introduce:

- data duplication
- risk of inconsistency
- complex backfill and correction logic
- additional write paths and race conditions

These states are functions of authoritative data such as payments,
due dates, and current time.

## Decision

Derived states (e.g. delinquency) are **computed dynamically at read time**
and are **not persisted** in the database.

The database remains the source of truth for:

- events (payments)
- immutable facts
- timestamps

Derived states are calculated using queries or application logic
based on the latest data.

## Consequences

### Positive

- Single source of truth
- No risk of stale or incorrect derived data
- Simpler write model
- Easier schema evolution

### Negative

- Slightly higher read-time computation cost
- Requires careful query and index design for performance
- Complex derivations may need optimization or caching later

## Notes

If a derived state becomes:

- extremely expensive to compute
- required at very high read frequency
- or needed for historical snapshots

We may revisit this decision and introduce:

- materialized views
- background projections
- or explicitly persisted state

Such a change should be captured in a new ADR.
