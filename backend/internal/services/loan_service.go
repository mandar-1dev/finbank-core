package services

import (
	"context"
	"database/sql"
	"math"

	"corebank/backend/internal/models"
	"corebank/backend/internal/util"
)

type LoanService struct {
	db    *sql.DB
	audit *AuditLogService
}

func NewLoanService(db *sql.DB, audit *AuditLogService) *LoanService {
	return &LoanService{db: db, audit: audit}
}

type LoanEstimate struct {
	MonthlyEMI     float64 `json:"monthlyEmi"`
	TotalInterest  float64 `json:"totalInterest"`
	TotalRepayment float64 `json:"totalRepayment"`
}

// CalculateEMI applies the standard reducing-balance EMI formula:
//
//	EMI = P * r * (1+r)^n / ((1+r)^n - 1)
//
// where P is principal, r is the monthly interest rate (annual rate / 12 /
// 100), and n is the tenure in months.
func CalculateEMI(principal, annualRatePercent float64, tenureMonths int) LoanEstimate {
	if tenureMonths <= 0 {
		return LoanEstimate{}
	}
	monthlyRate := annualRatePercent / 12 / 100

	var emi float64
	if monthlyRate == 0 {
		emi = principal / float64(tenureMonths)
	} else {
		factor := math.Pow(1+monthlyRate, float64(tenureMonths))
		emi = principal * monthlyRate * factor / (factor - 1)
	}

	total := emi * float64(tenureMonths)
	return LoanEstimate{
		MonthlyEMI:     round2(emi),
		TotalInterest:  round2(total - principal),
		TotalRepayment: round2(total),
	}
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func (s *LoanService) Apply(ctx context.Context, customerID, accountID int64, principal, annualRate float64, tenureMonths int) (models.Loan, error) {
	if principal <= 0 {
		return models.Loan{}, errInvalidAmount()
	}
	if tenureMonths <= 0 {
		return models.Loan{}, util.BadRequest("tenureMonths must be greater than zero")
	}
	estimate := CalculateEMI(principal, annualRate, tenureMonths)

	res, err := s.db.ExecContext(ctx, `
		INSERT INTO loans (customer_id, account_id, principal_amount, interest_rate, tenure_months, emi_amount, status)
		VALUES (?, ?, ?, ?, ?, ?, 'PENDING')`,
		customerID, accountID, principal, annualRate, tenureMonths, estimate.MonthlyEMI)
	if err != nil {
		return models.Loan{}, err
	}
	id, _ := res.LastInsertId()
	_ = s.audit.Record(ctx, s.db, "LOAN_APPLIED", "CUSTOMER", &customerID, "LOAN", &id,
		"Simulated loan application submitted")
	return s.Get(ctx, id)
}

func (s *LoanService) Get(ctx context.Context, id int64) (models.Loan, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, customer_id, account_id, principal_amount, interest_rate, tenure_months, emi_amount, status, applied_at
		FROM loans WHERE id = ?`, id)
	l, err := scanLoan(row)
	if err == sql.ErrNoRows {
		return models.Loan{}, errLoanNotFound()
	}
	return l, err
}

func (s *LoanService) ListByCustomer(ctx context.Context, customerID int64) ([]models.Loan, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, customer_id, account_id, principal_amount, interest_rate, tenure_months, emi_amount, status, applied_at
		FROM loans WHERE customer_id = ? ORDER BY id DESC`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Loan
	for rows.Next() {
		l, err := scanLoan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (s *LoanService) ListAll(ctx context.Context) ([]models.Loan, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, customer_id, account_id, principal_amount, interest_rate, tenure_months, emi_amount, status, applied_at
		FROM loans ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Loan
	for rows.Next() {
		l, err := scanLoan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func scanLoan(row rowScanner) (models.Loan, error) {
	var l models.Loan
	err := row.Scan(&l.ID, &l.CustomerID, &l.AccountID, &l.PrincipalAmount, &l.InterestRate, &l.TenureMonths, &l.EMIAmount, &l.Status, &l.AppliedAt)
	return l, err
}
