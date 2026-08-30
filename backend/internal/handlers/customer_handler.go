package handlers

import (
	"net/http"
	"strconv"

	"corebank/backend/internal/services"
	"corebank/backend/internal/util"
)

type CustomerHandler struct {
	customers     *services.CustomerService
	accounts      *services.AccountService
	transactions  *services.TransactionService
	beneficiaries *services.BeneficiaryService
	cards         *services.CardService
	loans         *services.LoanService
}

func NewCustomerHandler(
	customers *services.CustomerService,
	accounts *services.AccountService,
	transactions *services.TransactionService,
	beneficiaries *services.BeneficiaryService,
	cards *services.CardService,
	loans *services.LoanService,
) *CustomerHandler {
	return &CustomerHandler{customers, accounts, transactions, beneficiaries, cards, loans}
}

func pathID(r *http.Request, name string) (int64, error) {
	v := r.PathValue(name)
	id, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, util.BadRequest("invalid id: " + v)
	}
	return id, nil
}

func (h *CustomerHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.customers.List(r.Context())
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusOK, list)
}

func (h *CustomerHandler) Get(w http.ResponseWriter, r *http.Request) {
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
	util.WriteData(w, http.StatusOK, c)
}

func (h *CustomerHandler) Accounts(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		util.WriteError(w, err)
		return
	}
	if _, err := h.customers.Get(r.Context(), id); err != nil {
		util.WriteError(w, err)
		return
	}
	list, err := h.accounts.ListByCustomer(r.Context(), id)
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusOK, list)
}

func (h *CustomerHandler) Transactions(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		util.WriteError(w, err)
		return
	}
	list, err := h.transactions.ListByCustomer(r.Context(), id, 100)
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusOK, list)
}

func (h *CustomerHandler) Beneficiaries(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		util.WriteError(w, err)
		return
	}
	list, err := h.beneficiaries.ListByCustomer(r.Context(), id)
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusOK, list)
}

func (h *CustomerHandler) Cards(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		util.WriteError(w, err)
		return
	}
	list, err := h.cards.ListByCustomer(r.Context(), id)
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusOK, list)
}

func (h *CustomerHandler) Loans(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
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
