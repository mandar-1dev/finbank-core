package services

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

// These integration tests exercise the real MySQL transaction, locking,
// and constraint behaviour — the parts that matter most for financial
// correctness are exactly the parts that a mock can't validate honestly.
// They connect to a dedicated corebank_test_db (never corebank_db) and
// wipe/reseed its tables before every test.

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("COREBANK_TEST_DSN")
	if dsn == "" {
		dsn = "corebank:corebank_pass@tcp(127.0.0.1:3306)/corebank_test_db?parseTime=true&charset=utf8mb4"
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("opening test db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("skipping integration test: cannot reach MySQL test database (%v). Run `docker compose up mysql` or a local MySQL instance and set COREBANK_TEST_DSN.", err)
	}
	return db
}

// resetFixtures truncates every table and inserts two accounts (A with
// 10000, B with 5000) and one FROZEN account (C), returning their ids.
func resetFixtures(t *testing.T, db *sql.DB) (accountA, accountB, accountFrozen int64) {
	t.Helper()
	statements := []string{
		"SET FOREIGN_KEY_CHECKS=0",
		"TRUNCATE TABLE ledger_entries",
		"TRUNCATE TABLE transactions",
		"TRUNCATE TABLE audit_logs",
		"TRUNCATE TABLE accounts",
		"TRUNCATE TABLE customers",
		"SET FOREIGN_KEY_CHECKS=1",
		"INSERT INTO account_types (id, code, label) VALUES (1,'SAVINGS','Savings Account'),(2,'CURRENT','Current Account') ON DUPLICATE KEY UPDATE label=VALUES(label)",
		"INSERT INTO customers (id, customer_number, full_name, email, phone, date_of_birth, address) VALUES (1,'T1','Test A','a@test.local','1','2000-01-01','x'),(2,'T2','Test B','b@test.local','2','2000-01-01','x')",
	}
	for _, s := range statements {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("fixture setup failed on %q: %v", s, err)
		}
	}

	res, err := db.Exec(`INSERT INTO accounts (account_number, customer_id, account_type_id, balance, available_balance, status) VALUES (?,1,1,10000,10000,'ACTIVE')`, "9999000000000001")
	if err != nil {
		t.Fatal(err)
	}
	accountA, _ = res.LastInsertId()

	res, err = db.Exec(`INSERT INTO accounts (account_number, customer_id, account_type_id, balance, available_balance, status) VALUES (?,2,1,5000,5000,'ACTIVE')`, "9999000000000002")
	if err != nil {
		t.Fatal(err)
	}
	accountB, _ = res.LastInsertId()

	res, err = db.Exec(`INSERT INTO accounts (account_number, customer_id, account_type_id, balance, available_balance, status) VALUES (?,1,1,3000,3000,'FROZEN')`, "9999000000000003")
	if err != nil {
		t.Fatal(err)
	}
	accountFrozen, _ = res.LastInsertId()

	return
}
