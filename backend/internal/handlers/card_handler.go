package handlers

import (
	"net/http"

	"corebank/backend/internal/services"
	"corebank/backend/internal/util"
)

type CardHandler struct {
	cards *services.CardService
}

func NewCardHandler(c *services.CardService) *CardHandler {
	return &CardHandler{cards: c}
}

func (h *CardHandler) ListByCustomer(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "customerId")
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

func (h *CardHandler) Freeze(w http.ResponseWriter, r *http.Request) {
	h.setStatus(w, r, "FROZEN")
}

func (h *CardHandler) Unfreeze(w http.ResponseWriter, r *http.Request) {
	h.setStatus(w, r, "ACTIVE")
}

func (h *CardHandler) setStatus(w http.ResponseWriter, r *http.Request, status string) {
	id, err := pathID(r, "id")
	if err != nil {
		util.WriteError(w, err)
		return
	}
	c, err := h.cards.SetStatus(r.Context(), id, status)
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusOK, c)
}

type cardPaymentRequest struct {
	Merchant string  `json:"merchant"`
	Amount   float64 `json:"amount"`
}

func (h *CardHandler) Pay(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		util.WriteError(w, err)
		return
	}
	var req cardPaymentRequest
	if err := util.DecodeJSON(r, &req); err != nil {
		util.WriteError(w, err)
		return
	}
	txn, err := h.cards.Pay(r.Context(), id, req.Merchant, req.Amount)
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusCreated, txn)
}
