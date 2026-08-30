# CoreBank — Architecture

## Overview

CoreBank is a three-tier simulation:

```
Browser (vanilla JS)  →  Go HTTP API  →  MySQL 8
```

There is no authentication layer anywhere in the stack. The "Customer" and
"Banker" entry points on the landing page are simply two different UIs
against the same open API — a `selectedCustomerId` in `localStorage` is
the only piece of client-side state, and it exists purely to remember
which demo customer you're looking at.

## Backend layout

```
backend/
├── cmd/server/main.go        entrypoint: wiring + routing
└── internal/
    ├── config/                env-var loading
    ├── db/                    *sql.DB connection setup
    ├── models/                API-facing DTOs
    ├── services/               business logic (see below)
    ├── handlers/               HTTP layer: decode → call service → encode
    ├── middleware/              CORS, panic recovery, request logging
    └── util/                   HTTP helpers, masking, reference generation
```

The backend deliberately avoids an ORM. Every query that touches money is
written by hand with `database/sql`, so the exact SQL executed during a
transfer is visible in `internal/services/transfer_service.go` rather than
generated behind an abstraction. For a project meant to be explained line
by line in an interview, that trade-off is worth the extra boilerplate.

### Request flow

```
HTTP request
   │
   ▼
middleware (recover → CORS → logger)
   │
   ▼
handler   — decodes JSON, calls the service, encodes the response/error
   │
   ▼
service   — business rules, transactions, locking
   │
   ▼
database/sql → MySQL
```

Handlers never touch SQL. Services never touch `http.ResponseWriter`. This
keeps the money-movement logic testable without spinning up an HTTP
server — see `internal/services/*_test.go`, which call services directly
against a real test database.

## Money movement: deposit, withdraw, transfer

All three operations follow the same shape:

1. Open a database transaction (`db.BeginTx`).
2. `SELECT ... FOR UPDATE` the account row(s) involved — this takes a
   row-level lock that blocks any other transaction trying to read the
   same row until this one commits or rolls back.
3. Validate business rules against the *locked* balance (not a balance
   read before the lock was acquired).
4. `UPDATE` the balance(s).
5. Insert a `transactions` row and matching `ledger_entries` (double-entry:
   one DEBIT, one CREDIT for a transfer).
6. Insert an `audit_logs` row, in the same transaction.
7. Commit. If anything above returned an error, the deferred
   `tx.Rollback()` undoes everything — a failed step never leaves a
   half-completed transfer in the database.

### Concurrency

A transfer locks both the source and destination account. To avoid two
transfers deadlocking each other by locking the same two accounts in
opposite orders, `TransferService.Transfer` always acquires locks in
ascending account-ID order regardless of which account is the sender.
See `TestTransfer_ConcurrentOverdraft` in
`internal/services/transfer_service_test.go` for a test that fires two
simultaneous transfers exceeding a shared account's balance and asserts
exactly one succeeds and the balance never goes negative.

### Idempotency

`POST /api/transactions/transfer` accepts an optional `Idempotency-Key`
header. If a transaction with that key already exists, it's returned
as-is and no new transfer happens — safe to retry a request that timed
out or was double-submitted by an eager double-click.

## Suspicious transaction simulation

`classifyRisk` in `transfer_service.go` is a simple, fully documented rule
set (absolute amount thresholds + percentage of sender's balance) — not a
real fraud model. Transfers above the thresholds get `risk_level =
MEDIUM/HIGH` and an additional `SUSPICIOUS_TRANSACTION_FLAGGED` audit
entry, both viewable from the banker dashboard.

## Frontend layout

Vanilla HTML/CSS/JS, no build step, no framework. `js/api.js` is the only
file that calls `fetch()` — every page-specific script (`transfer.js`,
`cards.js`, ...) goes through it, so switching the API base URL or adding
a global error handler happens in one place.
