// Package models holds the plain data structures returned by the API.
// These are DTO-shaped on purpose (not a 1:1 mirror of the tables) so the
// database schema can change without breaking the frontend contract.
package models

import "time"

type Customer struct {
	ID             int64     `json:"id"`
	CustomerNumber string    `json:"customerNumber"`
	FullName       string    `json:"fullName"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	DateOfBirth    string    `json:"dateOfBirth"`
	Address        string    `json:"address"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
}

type Account struct {
	ID               int64     `json:"id"`
	AccountNumber    string    `json:"accountNumber"`
	MaskedNumber     string    `json:"maskedNumber"`
	CustomerID       int64     `json:"customerId"`
	AccountTypeCode  string    `json:"accountType"`
	Balance          float64   `json:"balance"`
	AvailableBalance float64   `json:"availableBalance"`
	Status           string    `json:"status"`
	OpenedAt         time.Time `json:"openedAt"`
}

type Transaction struct {
	ID                    int64     `json:"id"`
	TransactionRef        string    `json:"transactionRef"`
	Type                  string    `json:"type"`
	Status                string    `json:"status"`
	SourceAccountID       *int64    `json:"sourceAccountId,omitempty"`
	DestinationAccountID  *int64    `json:"destinationAccountId,omitempty"`
	SourceAccountNumber   *string   `json:"sourceAccountNumber,omitempty"`
	DestinationAccountNum *string   `json:"destinationAccountNumber,omitempty"`
	Amount                float64   `json:"amount"`
	Description           string    `json:"description"`
	RiskLevel             string    `json:"riskLevel"`
	CreatedAt             time.Time `json:"createdAt"`
}

type LedgerEntry struct {
	ID            int64     `json:"id"`
	TransactionID int64     `json:"transactionId"`
	AccountID     int64     `json:"accountId"`
	EntryType     string    `json:"entryType"`
	Amount        float64   `json:"amount"`
	CreatedAt     time.Time `json:"createdAt"`
}

type Beneficiary struct {
	ID            int64     `json:"id"`
	CustomerID    int64     `json:"customerId"`
	Name          string    `json:"name"`
	AccountNumber string    `json:"accountNumber"`
	IFSCCode      string    `json:"ifscCode"`
	Nickname      string    `json:"nickname"`
	CreatedAt     time.Time `json:"createdAt"`
}

type Card struct {
	ID            int64   `json:"id"`
	AccountID     int64   `json:"accountId"`
	MaskedNumber  string  `json:"maskedNumber"`
	CardHolder    string  `json:"cardHolder"`
	ExpiryMonth   int     `json:"expiryMonth"`
	ExpiryYear    int     `json:"expiryYear"`
	Status        string  `json:"status"`
	SpendingLimit float64 `json:"spendingLimit"`
}

type Loan struct {
	ID              int64     `json:"id"`
	CustomerID      int64     `json:"customerId"`
	AccountID       int64     `json:"accountId"`
	PrincipalAmount float64   `json:"principalAmount"`
	InterestRate    float64   `json:"interestRate"`
	TenureMonths    int       `json:"tenureMonths"`
	EMIAmount       float64   `json:"emiAmount"`
	Status          string    `json:"status"`
	AppliedAt       time.Time `json:"appliedAt"`
}

type AuditLog struct {
	ID          int64     `json:"id"`
	Action      string    `json:"action"`
	ActorType   string    `json:"actorType"`
	ActorID     *int64    `json:"actorId,omitempty"`
	Entity      string    `json:"entity"`
	EntityID    *int64    `json:"entityId,omitempty"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

type DailyVolume struct {
	Date   string  `json:"date"`
	Count  int     `json:"count"`
	Volume float64 `json:"volume"`
}

type DashboardStats struct {
	TotalCustomers         int     `json:"totalCustomers"`
	TotalAccounts          int     `json:"totalAccounts"`
	ActiveAccounts         int     `json:"activeAccounts"`
	FrozenAccounts         int     `json:"frozenAccounts"`
	TransactionsToday      int     `json:"transactionsToday"`
	TransactionVolumeToday float64 `json:"transactionVolumeToday"`
	ActiveLoans            int     `json:"activeLoans"`
	PendingLoans           int     `json:"pendingLoans"`
	SuspiciousTransactions int     `json:"suspiciousTransactions"`
}
