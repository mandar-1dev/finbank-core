package handlers

import (
	"net/http"

	"corebank/backend/internal/services"
	"corebank/backend/internal/util"
)

type AccountHandler struct {
	accounts     *services.AccountService
	transactions *services.TransactionService
}

func NewAccountHandler(accounts *services.AccountService, transactions *services.TransactionService) *AccountHandler {
	return &AccountHandler{accounts, transactions}
}

func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.accounts.List(r.Context())
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusOK, list)
}

func (h *AccountHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
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

func (h *AccountHandler) Transactions(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		util.WriteError(w, err)
		return
	}
	if _, err := h.accounts.Get(r.Context(), id); err != nil {
		util.WriteError(w, err)
		return
	}
	list, err := h.transactions.ListByAccount(r.Context(), id, 100)
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusOK, list)
}
