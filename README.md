# CoreBank

A full-stack **core banking simulation platform** built to demonstrate
financial transaction processing, database integrity, backend
architecture, and banking operations.

> **Disclaimer:** CoreBank is an educational simulation. It does not
> connect to real banks, real payment systems, real cards, or real
> financial accounts. Every customer, account, and balance in this
> project is fictional and lives only in a local MySQL database.

## What this is

CoreBank lets you pick between two simulated experiences from a single
landing page:

- **Customer mode** — select one of six predefined demo customers and
  simulate real banking operations: deposits, withdrawals, transfers,
  beneficiaries, debit cards, and loans.
- **Banker mode** — an internal operations dashboard for managing
  customers, freezing/unfreezing accounts, reviewing transactions,
  monitoring flagged activity, and reading the audit trail.

There is **no login system anywhere**. Both entry points are just two
different UIs against the same open API; the "selected customer" is a
convenience stored in `localStorage`, never a credential.

## Features

- Deposit / withdraw / transfer with full **ACID** guarantees
- **Double-entry ledger** — every transfer produces a matching DEBIT and
  CREDIT row in `ledger_entries`
- **Row-level locking** (`SELECT ... FOR UPDATE`) with fixed lock ordering
  to make concurrent transfers on the same accounts deadlock-free and
  overdraft-proof — see [`docs/architecture.md`](docs/architecture.md)
- **Idempotency keys** on transfers, so a retried request never moves
  money twice
- Simulated debit cards (freeze/unfreeze, spending limits, payment sim)
- Loan EMI calculator using the standard reducing-balance formula, plus a
  simulated application flow
- Simple, documented **rule-based suspicious-transaction flagging**
  (not real fraud detection)
- Full **audit log** of every customer and banker action
- Banker dashboard with Chart.js visualizations

## Tech stack

| Layer | Technology |
|---|---|
| Frontend | HTML5, CSS3 (Grid/Flexbox/variables), vanilla JavaScript, Chart.js |
| Backend | Go (standard library `net/http`, `database/sql`) |
| Database | MySQL 8 |
| Testing | Go's built-in `testing` package |
| Containerization | Docker, Docker Compose |

No frontend framework, no ORM, no authentication library — deliberately.
The goal is a project a second-year CS student can read end-to-end and
explain in an interview, not a showcase of every tool available.

## Folder structure

```
corebank/
├── frontend/            HTML/CSS/JS — no build step
│   ├── *.html            customer-facing pages
│   ├── banker/            banker-facing pages
│   ├── css/ js/            shared styles and scripts
├── backend/              Go API
│   ├── cmd/server/         entrypoint
│   └── internal/           config, db, models, services, handlers, middleware, util
├── database/             schema.sql, seed.sql
├── docs/                  architecture.md, database.md, api.md
├── postman/               CoreBank.postman_collection.json
├── Dockerfile / docker-compose.yml
└── .env.example
```

## Database design

See [`docs/database.md`](docs/database.md) for the full schema
walkthrough. Ten tables: `customers`, `account_types`, `accounts`,
`transactions`, `ledger_entries`, `beneficiaries`, `cards`, `loans`,
`loan_payments`, `audit_logs`.

## Customer flow

```
Home → Customer → Customer Selection → Customer Dashboard
                                          → Accounts / Transfer / Transactions
                                          → Beneficiaries / Cards / Loans / Profile
```

## Banker flow

```
Home → Banker → Banker Dashboard
                   → Customers → Customer Details (freeze/unfreeze accounts)
                   → Accounts / Transactions / Suspicious Activity
                   → Cards / Loans / Audit Logs
```

## Demo customers

Seeded in `database/seed.sql`:

| Name | Accounts |
|---|---|
| Rahul Sharma | Savings ****4821, Current ****9034 |
| Priya Mehta | Savings ****7312 |
| Arjun Patil | Current ****1945 |
| Sneha Kulkarni | Savings ****5567 |
| Aditya Deshmukh | Savings ****6678 |
| Neha Joshi | Current ****8890 *(pre-frozen, for demoing banker unfreeze)* |

## How to run

### Option A — Docker (recommended)

```bash
docker compose up
```

This starts MySQL (auto-loading `schema.sql` and `seed.sql` on first
boot) and the Go backend on `http://localhost:8080`. Then serve the
frontend with any static file server, e.g.:

```bash
cd frontend && python3 -m http.server 5500
# open http://localhost:5500
```

### Option B — Manual local setup

**1. MySQL**

```bash
mysql -u root -p < database/schema.sql
mysql -u root -p < database/seed.sql

mysql -u root -p -e "
  CREATE USER 'corebank'@'localhost' IDENTIFIED WITH mysql_native_password BY 'corebank_pass';
  GRANT ALL PRIVILEGES ON corebank_db.* TO 'corebank'@'localhost';
  FLUSH PRIVILEGES;"
```

> The bundled `go-sql-driver/mysql` version used here needs
> `mysql_native_password` auth (MySQL 8's default,
> `caching_sha2_password`, isn't supported by that driver version without
> an extra dependency) — the command above sets that up for you.

**2. Backend**

```bash
cd backend
cp ../.env.example .env   # optional — defaults already match the setup above
go run ./cmd/server
```

The API starts on `http://localhost:8080`. Check `GET /api/health`.

**3. Frontend**

```bash
cd frontend
python3 -m http.server 5500
# open http://localhost:5500 in a browser
```

No build step — it's static HTML/CSS/JS.

## Environment variables

| Variable | Default |
|---|---|
| `DB_HOST` | `127.0.0.1` |
| `DB_PORT` | `3306` |
| `DB_NAME` | `corebank_db` |
| `DB_USERNAME` | `corebank` |
| `DB_PASSWORD` | `corebank_pass` |
| `SERVER_PORT` | `8080` |

See `.env.example`.

## API documentation

Full endpoint reference: [`docs/api.md`](docs/api.md). Importable Postman
collection: [`postman/CoreBank.postman_collection.json`](postman/CoreBank.postman_collection.json).

## Testing

```bash
cd backend

# unit tests only (no database needed)
go test ./internal/util/...

# full suite, including integration tests against a real MySQL database
mysql -u root -e "CREATE DATABASE IF NOT EXISTS corebank_test_db;"
mysql -u root corebank_test_db < ../database/schema.sql   # or re-run schema.sql with the db name swapped
COREBANK_TEST_DSN="corebank:corebank_pass@tcp(127.0.0.1:3306)/corebank_test_db?parseTime=true&charset=utf8mb4" \
  go test ./... -v
```

The integration suite covers deposit, withdrawal, insufficient balance,
frozen-account rejection, same-account transfer rejection, idempotent
retry, and — the one that matters most — a concurrency test that fires
two simultaneous transfers exceeding a shared account's balance and
asserts exactly one succeeds and the balance never goes negative
(`TestTransfer_ConcurrentOverdraft`). All 20 tests pass; that
concurrency test was additionally run 5 times back-to-back with no
flakes during development.

## Future improvements

Ideas intentionally left out of this simulation to keep it a readable,
single-weekend-scale project:

- Real double-entry general ledger with a chart of accounts (this project
  ties ledger entries directly to accounts, not to ledger account codes)
- Scheduled/recurring transfers
- Statement (PDF) generation
- Rate limiting and request throttling
- A proper fraud-scoring model in place of the simple threshold rules
- Multi-currency support

## License

MIT — see [`LICENSE`](LICENSE).
