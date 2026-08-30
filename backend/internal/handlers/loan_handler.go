package handlers

import (
	"net/http"

	"corebank/backend/internal/services"
	"corebank/backend/internal/util"
)

type LoanHandler struct {
	loans *services.LoanService
}

func NewLoanHandler(l *services.LoanService) *LoanHandler {
	return &LoanHandler{loans: l}
}

func (h *LoanHandler) ListByCustomer(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "customerId")
	if err != nil {
		util.WriteError(w, err)
		return
	}
	list, err := h.loans.ListByCustomer(r.Context(), id)
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusOK, list)
}

type loanEstimateRequest struct {
	Principal    float64 `json:"principal"`
	InterestRate float64 `json:"interestRate"`
	TenureMonths int     `json:"tenureMonths"`
}

func (h *LoanHandler) Estimate(w http.ResponseWriter, r *http.Request) {
	var req loanEstimateRequest
	if err := util.DecodeJSON(r, &req); err != nil {
		util.WriteError(w, err)
		return
	}
	if req.Principal <= 0 || req.TenureMonths <= 0 {
		util.WriteError(w, util.BadRequest("principal and tenureMonths must be greater than zero"))
		return
	}
	estimate := services.CalculateEMI(req.Principal, req.InterestRate, req.TenureMonths)
	util.WriteData(w, http.StatusOK, estimate)
}

type loanApplyRequest struct {
	CustomerID   int64   `json:"customerId"`
	AccountID    int64   `json:"accountId"`
	Principal    float64 `json:"principal"`
	InterestRate float64 `json:"interestRate"`
	TenureMonths int     `json:"tenureMonths"`
}

func (h *LoanHandler) Apply(w http.ResponseWriter, r *http.Request) {
	var req loanApplyRequest
	if err := util.DecodeJSON(r, &req); err != nil {
		util.WriteError(w, err)
		return
	}
	loan, err := h.loans.Apply(r.Context(), req.CustomerID, req.AccountID, req.Principal, req.InterestRate, req.TenureMonths)
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusCreated, loan)
}
