package handlers

import (
	"net/http"

	"corebank/backend/internal/services"
	"corebank/backend/internal/util"
)

type BankerHandler struct {
	banker       *services.BankerService
	customers    *services.CustomerService
	accounts     *services.AccountService
	transactions *services.TransactionService
	cards        *services.CardService
	loans        *services.LoanService
	audit        *services.AuditLogService
}

func NewBankerHandler(
	banker *services.BankerService,
	customers *services.CustomerService,
	accounts *services.AccountService,
	transactions *services.TransactionService,
	cards *services.CardService,
	loans *services.LoanService,
	audit *services.AuditLogService,
) *BankerHandler {
	return &BankerHandler{banker, customers, accounts, transactions, cards, loans, audit}
}

func (h *BankerHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := h.banker.DashboardStats(r.Context())
	if err != nil {
		util.WriteError(w, err)
		return
	}
	accountTypes, err := h.banker.AccountTypeDistribution(r.Context())
	if err != nil {
		util.WriteError(w, err)
		return
	}
	loanStatus, err := h.banker.LoanStatusDistribution(r.Context())
	if err != nil {
		util.WriteError(w, err)
		return
	}
	volume, err := h.banker.DailyTransactionVolume(r.Context(), 14)
	if err != nil {
		util.WriteError(w, err)
		return
	}

	util.WriteData(w, http.StatusOK, map[string]interface{}{
		"stats":                   stats,
		"accountTypeDistribution": accountTypes,
		"loanStatusDistribution":  loanStatus,
		"dailyVolume":             volume,
	})
}

func (h *BankerHandler) Customers(w http.ResponseWriter, r *http.Request) {
	list, err := h.customers.List(r.Context())
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusOK, list)
}

func (h *BankerHandler) CustomerDetails(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		util.WriteError(w, err)
		return
	}
	c, err := h.customers.Get(r.Context(), id)
	if err != nil {
		util.WriteError(w, err)
		return
	}
	accounts, err := h.accounts.ListByCustomer(r.Context(), id)
	if err != nil {
		util.WriteError(w, err)
		return
	}
	cards, err := h.cards.ListByCustomer(r.Context(), id)
	if err != nil {
		util.WriteError(w, err)
		return
	}
	loans, err := h.loans.ListByCustomer(r.Context(), id)
	if err != nil {
		util.WriteError(w, err)
		return
	}
	transactions, err := h.transactions.ListByCustomer(r.Context(), id, 50)
	if err != nil {
		util.WriteError(w, err)
		return
	}

	util.WriteData(w, http.StatusOK, map[string]interface{}{
		"customer":     c,
		"accounts":     accounts,
		"cards":        cards,
		"loans":        loans,
		"transactions": transactions,
	})
}

func (h *BankerHandler) Accounts(w http.ResponseWriter, r *http.Request) {
	list, err := h.accounts.List(r.Context())
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusOK, list)
}

func (h *BankerHandler) Transactions(w http.ResponseWriter, r *http.Request) {
	list, err := h.transactions.ListAll(r.Context(), 200)
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusOK, list)
}

func (h *BankerHandler) SuspiciousTransactions(w http.ResponseWriter, r *http.Request) {
	list, err := h.transactions.ListSuspicious(r.Context(), 200)
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusOK, list)
}

func (h *BankerHandler) Cards(w http.ResponseWriter, r *http.Request) {
	list, err := h.cards.ListAll(r.Context())
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusOK, list)
}

func (h *BankerHandler) Loans(w http.ResponseWriter, r *http.Request) {
	list, err := h.loans.ListAll(r.Context())
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusOK, list)
}

func (h *BankerHandler) AuditLogs(w http.ResponseWriter, r *http.Request) {
	list, err := h.audit.List(r.Context(), 200)
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusOK, list)
}

func (h *BankerHandler) FreezeAccount(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		util.WriteError(w, err)
		return
	}
	if err := h.banker.FreezeAccount(r.Context(), id); err != nil {
		util.WriteError(w, err)
		return
	}
	a, err := h.accounts.Get(r.Context(), id)
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusOK, a)
}

func (h *BankerHandler) UnfreezeAccount(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		util.WriteError(w, err)
		return
	}
	if err := h.banker.UnfreezeAccount(r.Context(), id); err != nil {
		util.WriteError(w, err)
		return
	}
	a, err := h.accounts.Get(r.Context(), id)
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusOK, a)
}
