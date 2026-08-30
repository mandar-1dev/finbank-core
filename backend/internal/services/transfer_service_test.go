package services

import (
	"context"
	"sync"
	"testing"
)

func TestTransfer_Success(t *testing.T) {
	db := testDB(t)
	audit := NewAuditLogService(db)
	transferSvc := NewTransferService(db, audit)
	accountA, accountB, _ := resetFixtures(t, db)

	txn, err := transferSvc.Transfer(context.Background(), accountA, accountB, 2000, "test transfer", "")
	if err != nil {
		t.Fatalf("transfer failed: %v", err)
	}
	if txn.Type != "TRANSFER" || txn.Status != "SUCCESS" {
		t.Errorf("unexpected transaction: %+v", txn)
	}

	var balA, balB float64
	db.QueryRow("SELECT balance FROM accounts WHERE id = ?", accountA).Scan(&balA)
	db.QueryRow("SELECT balance FROM accounts WHERE id = ?", accountB).Scan(&balB)
	if balA != 8000 {
		t.Errorf("sender balance should be 8000, got %.2f", balA)
	}
	if balB != 7000 {
		t.Errorf("receiver balance should be 7000, got %.2f", balB)
	}

	// exactly one DEBIT and one CREDIT ledger entry must exist for this txn
	var ledgerCount int
	db.QueryRow("SELECT COUNT(*) FROM ledger_entries WHERE transaction_id = ?", txn.ID).Scan(&ledgerCount)
	if ledgerCount != 2 {
		t.Errorf("expected exactly 2 ledger entries (debit+credit), got %d", ledgerCount)
	}
}

func TestTransfer_InsufficientBalance_NoPartialMovement(t *testing.T) {
	db := testDB(t)
	audit := NewAuditLogService(db)
	transferSvc := NewTransferService(db, audit)
	accountA, accountB, _ := resetFixtures(t, db)

	_, err := transferSvc.Transfer(context.Background(), accountA, accountB, 999999, "too much", "")
	if err == nil {
		t.Fatal("expected insufficient balance error, got nil")
	}

	// Neither balance should have moved even a rupee — this is the core
	// "sender loses money but receiver never gets it" failure mode the
	// spec explicitly calls out, and it must be impossible.
	var balA, balB float64
	db.QueryRow("SELECT balance FROM accounts WHERE id = ?", accountA).Scan(&balA)
	db.QueryRow("SELECT balance FROM accounts WHERE id = ?", accountB).Scan(&balB)
	if balA != 10000 || balB != 5000 {
		t.Errorf("balances must be untouched on a failed transfer, got A=%.2f B=%.2f", balA, balB)
	}
}

func TestTransfer_FrozenSenderRejected(t *testing.T) {
	db := testDB(t)
	audit := NewAuditLogService(db)
	transferSvc := NewTransferService(db, audit)
	_, accountB, frozen := resetFixtures(t, db)

	_, err := transferSvc.Transfer(context.Background(), frozen, accountB, 100, "should fail", "")
	if err == nil {
		t.Fatal("expected error transferring from a frozen account, got nil")
	}
}

func TestTransfer_SameAccountRejected(t *testing.T) {
	db := testDB(t)
	audit := NewAuditLogService(db)
	transferSvc := NewTransferService(db, audit)
	accountA, _, _ := resetFixtures(t, db)

	_, err := transferSvc.Transfer(context.Background(), accountA, accountA, 100, "self transfer", "")
	if err == nil {
		t.Fatal("expected error transferring an account to itself, got nil")
	}
}

func TestTransfer_IdempotencyKeyPreventsDoubleTransfer(t *testing.T) {
	db := testDB(t)
	audit := NewAuditLogService(db)
	transferSvc := NewTransferService(db, audit)
	accountA, accountB, _ := resetFixtures(t, db)

	key := "test-idempotency-key-001"
	first, err := transferSvc.Transfer(context.Background(), accountA, accountB, 1000, "first attempt", key)
	if err != nil {
		t.Fatalf("first transfer failed: %v", err)
	}

	// Simulate a retried request (double click / network retry) with the
	// same idempotency key.
	second, err := transferSvc.Transfer(context.Background(), accountA, accountB, 1000, "retry", key)
	if err != nil {
		t.Fatalf("retried transfer with same idempotency key should not error: %v", err)
	}
	if first.ID != second.ID {
		t.Errorf("expected the same transaction to be returned, got ids %d and %d", first.ID, second.ID)
	}

	var balA float64
	db.QueryRow("SELECT balance FROM accounts WHERE id = ?", accountA).Scan(&balA)
	if balA != 9000 { // only debited once: 10000 - 1000
		t.Errorf("expected balance 9000 after a single effective transfer, got %.2f (money moved twice!)", balA)
	}
}

// TestTransfer_ConcurrentOverdraft is the most important test in this
// project. It fires two simultaneous transfers, each larger than half of
// the sender's balance, from the same account. Without correct locking,
// both could read the pre-transfer balance, both could pass the balance
// check, and both could commit — silently taking the account negative.
// With SELECT ... FOR UPDATE inside a single database transaction, the
// second transfer must always see the first transfer's already-debited
// balance and correctly fail.
func TestTransfer_ConcurrentOverdraft(t *testing.T) {
	db := testDB(t)
	audit := NewAuditLogService(db)
	transferSvc := NewTransferService(db, audit)
	accountA, accountB, _ := resetFixtures(t, db) // A starts at 10000

	var wg sync.WaitGroup
	results := make([]error, 2)

	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func(idx int) {
			defer wg.Done()
			_, err := transferSvc.Transfer(context.Background(), accountA, accountB, 7000, "race", "")
			results[idx] = err
		}(i)
	}
	wg.Wait()

	successCount := 0
	for _, err := range results {
		if err == nil {
			successCount++
		}
	}
	if successCount != 1 {
		t.Fatalf("expected exactly 1 of 2 racing transfers to succeed, got %d successes", successCount)
	}

	var balA float64
	db.QueryRow("SELECT balance FROM accounts WHERE id = ?", accountA).Scan(&balA)
	if balA < 0 {
		t.Fatalf("account went negative: %.2f — concurrency control failed", balA)
	}
	if balA != 3000 { // exactly one 7000 transfer should have gone through
		t.Errorf("expected final balance 3000 after exactly one successful transfer, got %.2f", balA)
	}
}
