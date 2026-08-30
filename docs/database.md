# CoreBank — Database Design

Database: `corebank_db` (MySQL 8, InnoDB, utf8mb4).

## Entity overview

```
customers ──< accounts >── account_types
                │
                ├──< transactions (source_account_id / destination_account_id)
                │        │
                │        └──< ledger_entries
                │
                ├──< cards
                └──< loans ──< loan_payments

customers ──< beneficiaries

audit_logs   (standalone — references entities by entity/entity_id, not FK)
```

## Why a separate `ledger_entries` table

`accounts.balance` is the fast-path number the UI reads. `ledger_entries`
is the audit trail: every balance change is also recorded as a DEBIT or
CREDIT row tied to the transaction that caused it. A transfer produces
exactly two ledger rows (one DEBIT on the sender, one CREDIT on the
receiver) that always sum to zero — the classic double-entry invariant.
This means the balance on `accounts` is always reconstructable by summing
`ledger_entries` for that account, which is the property a real ledger
system is built around.

## Money as DECIMAL, never FLOAT

Every monetary column is `DECIMAL(18,2)`, and the Go layer uses
`float64` only at the JSON-encoding boundary — all comparisons and
arithmetic that decide whether a transfer succeeds happen in SQL against
the DECIMAL column, and MySQL's DECIMAL arithmetic is exact (no binary
floating-point rounding). `float64` in Go DTOs is a readability trade-off
appropriate for a two-decimal-place teaching project; a production system
would carry a fixed-point or minor-units integer type all the way through
the Go layer as well.

## Status enums

- `accounts.status`: `ACTIVE | FROZEN | CLOSED`
- `cards.status`: `ACTIVE | FROZEN | CLOSED`
- `transactions.status`: `PENDING | SUCCESS | FAILED | REVERSED`
- `transactions.risk_level`: `LOW | MEDIUM | HIGH`
- `loans.status`: `PENDING | APPROVED | REJECTED | ACTIVE | CLOSED`

## Constraints worth noting

- `accounts.balance` has a `CHECK (balance >= 0)` — belt-and-suspenders on
  top of the application-level balance check before every debit.
- `transactions.idempotency_key` is `UNIQUE` (nullable) — the database
  itself refuses a second row with the same key, which is what makes the
  idempotency guarantee hold even under a race (see
  `TestTransfer_IdempotencyKeyPreventsDoubleTransfer`).
- `transactions.transaction_ref` is `UNIQUE` — a human-friendly reference
  like `TXN-9F3K2A7Q1B`, separate from the numeric primary key.

## Seed data

`database/seed.sql` creates 6 fictional customers, 7 accounts (one
pre-frozen for demoing banker actions), 4 beneficiaries, 3 cards, 2 loans,
and a handful of sample transactions with matching ledger entries — the
project is usable immediately after `schema.sql` + `seed.sql`, no manual
data entry required.
