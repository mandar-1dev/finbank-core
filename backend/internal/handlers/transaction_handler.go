package handlers

import (
	"net/http"

	"corebank/backend/internal/services"
	"corebank/backend/internal/util"
)

type TransactionHandler struct {
	transactions *services.TransactionService
	transfers    *services.TransferService
}

func NewTransactionHandler(transactions *services.TransactionService, transfers *services.TransferService) *TransactionHandler {
	return &TransactionHandler{transactions, transfers}
}

type depositRequest struct {
	AccountID   int64   `json:"accountId"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}

func (h *TransactionHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	var req depositRequest
	if err := util.DecodeJSON(r, &req); err != nil {
		util.WriteError(w, err)
		return
	}
	txn, err := h.transactions.Deposit(r.Context(), req.AccountID, req.Amount, req.Description)
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusCreated, txn)
}

type withdrawRequest struct {
	AccountID   int64   `json:"accountId"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}

func (h *TransactionHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	var req withdrawRequest
	if err := util.DecodeJSON(r, &req); err != nil {
		util.WriteError(w, err)
		return
	}
	txn, err := h.transactions.Withdraw(r.Context(), req.AccountID, req.Amount, req.Description)
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusCreated, txn)
}

type transferRequest struct {
	FromAccountID int64   `json:"fromAccountId"`
	ToAccountID   int64   `json:"toAccountId"`
	Amount        float64 `json:"amount"`
	Description   string  `json:"description"`
}

func (h *TransactionHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req transferRequest
	if err := util.DecodeJSON(r, &req); err != nil {
		util.WriteError(w, err)
		return
	}
	// The Idempotency-Key header lets a retried request (double click,
	// network retry) return the original result instead of transferring
	// money twice. It's optional — if omitted, the request is processed
	// as a new transfer every time.
	key := r.Header.Get("Idempotency-Key")

	txn, err := h.transfers.Transfer(r.Context(), req.FromAccountID, req.ToAccountID, req.Amount, req.Description, key)
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusCreated, txn)
}

func (h *TransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		util.WriteError(w, err)
		return
	}
	txn, err := h.transactions.GetByID(r.Context(), id, "")
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusOK, txn)
}
