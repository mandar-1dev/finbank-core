package services

import (
	"context"
	"testing"
)

func TestDeposit_Success(t *testing.T) {
	db := testDB(t)
	audit := NewAuditLogService(db)
	txnSvc := NewTransactionService(db, audit)
	accountA, _, _ := resetFixtures(t, db)

	txn, err := txnSvc.Deposit(context.Background(), accountA, 1500, "test deposit")
	if err != nil {
		t.Fatalf("deposit failed: %v", err)
	}
	if txn.Amount != 1500 || txn.Type != "DEPOSIT" || txn.Status != "SUCCESS" {
		t.Errorf("unexpected transaction shape: %+v", txn)
	}

	var balance float64
	db.QueryRow("SELECT balance FROM accounts WHERE id = ?", accountA).Scan(&balance)
	if balance != 11500 {
		t.Errorf("expected balance 11500, got %.2f", balance)
	}
}

func TestDeposit_RejectsZeroAmount(t *testing.T) {
	db := testDB(t)
	audit := NewAuditLogService(db)
	txnSvc := NewTransactionService(db, audit)
	accountA, _, _ := resetFixtures(t, db)

	_, err := txnSvc.Deposit(context.Background(), accountA, 0, "invalid")
	if err == nil {
		t.Fatal("expected error for zero-amount deposit, got nil")
	}
}

func TestWithdraw_Success(t *testing.T) {
	db := testDB(t)
	audit := NewAuditLogService(db)
	txnSvc := NewTransactionService(db, audit)
	accountA, _, _ := resetFixtures(t, db)

	_, err := txnSvc.Withdraw(context.Background(), accountA, 4000, "test withdrawal")
	if err != nil {
		t.Fatalf("withdraw failed: %v", err)
	}

	var balance float64
	db.QueryRow("SELECT balance FROM accounts WHERE id = ?", accountA).Scan(&balance)
	if balance != 6000 {
		t.Errorf("expected balance 6000, got %.2f", balance)
	}
}

func TestWithdraw_InsufficientBalance(t *testing.T) {
	db := testDB(t)
	audit := NewAuditLogService(db)
	txnSvc := NewTransactionService(db, audit)
	accountA, _, _ := resetFixtures(t, db)

	_, err := txnSvc.Withdraw(context.Background(), accountA, 999999, "too much")
	if err == nil {
		t.Fatal("expected insufficient balance error, got nil")
	}

	// balance must be unchanged after a failed withdrawal
	var balance float64
	db.QueryRow("SELECT balance FROM accounts WHERE id = ?", accountA).Scan(&balance)
	if balance != 10000 {
		t.Errorf("balance should be untouched at 10000, got %.2f", balance)
	}
}

func TestWithdraw_FrozenAccountRejected(t *testing.T) {
	db := testDB(t)
	audit := NewAuditLogService(db)
	txnSvc := NewTransactionService(db, audit)
	_, _, frozen := resetFixtures(t, db)

	_, err := txnSvc.Withdraw(context.Background(), frozen, 100, "should fail")
	if err == nil {
		t.Fatal("expected error withdrawing from a frozen account, got nil")
	}
}

func TestDeposit_AllowedOnFrozenAccount(t *testing.T) {
	// Per spec: a freeze blocks money leaving an account, not entering it.
	db := testDB(t)
	audit := NewAuditLogService(db)
	txnSvc := NewTransactionService(db, audit)
	_, _, frozen := resetFixtures(t, db)

	_, err := txnSvc.Deposit(context.Background(), frozen, 500, "deposit into frozen account")
	if err != nil {
		t.Fatalf("expected deposit into frozen account to succeed, got: %v", err)
	}
}
