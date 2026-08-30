package services

import (
	"math"
	"testing"
)

func TestCalculateEMI_StandardLoan(t *testing.T) {
	// 300000 at 10.5% for 24 months — verified against a standard reducing
	// balance EMI calculator.
	est := CalculateEMI(300000, 10.5, 24)

	if math.Abs(est.MonthlyEMI-13912.81) > 1.0 {
		t.Errorf("expected EMI close to 13912.81, got %.2f", est.MonthlyEMI)
	}
	if est.TotalRepayment <= 300000 {
		t.Errorf("total repayment should exceed principal, got %.2f", est.TotalRepayment)
	}
	// total repayment must equal EMI * tenure (within rounding)
	expectedTotal := est.MonthlyEMI * 24
	if math.Abs(est.TotalRepayment-expectedTotal) > 1.0 {
		t.Errorf("total repayment %.2f should equal EMI*tenure %.2f", est.TotalRepayment, expectedTotal)
	}
	if math.Abs(est.TotalInterest-(est.TotalRepayment-300000)) > 0.01 {
		t.Errorf("total interest should equal total repayment minus principal")
	}
}

func TestCalculateEMI_ZeroInterest(t *testing.T) {
	est := CalculateEMI(120000, 0, 12)
	if est.MonthlyEMI != 10000 {
		t.Errorf("expected EMI of exactly 10000 for 0%% interest, got %.2f", est.MonthlyEMI)
	}
	if est.TotalInterest != 0 {
		t.Errorf("expected zero interest, got %.2f", est.TotalInterest)
	}
}

func TestCalculateEMI_InvalidTenure(t *testing.T) {
	est := CalculateEMI(100000, 10, 0)
	if est.MonthlyEMI != 0 {
		t.Errorf("expected zero-value estimate for invalid tenure, got %+v", est)
	}
}

func TestClassifyRisk(t *testing.T) {
	cases := []struct {
		amount, balanceBefore float64
		want                  string
	}{
		{1000, 100000, "LOW"},
		{80000, 200000, "MEDIUM"}, // above medium absolute threshold
		{250000, 500000, "HIGH"},  // above high absolute threshold
		{40000, 50000, "MEDIUM"},  // 80% of balance, below absolute thresholds
		{100, 0, "LOW"},           // guard against divide-by-zero
	}
	for _, c := range cases {
		got := classifyRisk(c.amount, c.balanceBefore)
		if got != c.want {
			t.Errorf("classifyRisk(%.0f, %.0f) = %s, want %s", c.amount, c.balanceBefore, got, c.want)
		}
	}
}
