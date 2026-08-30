package handlers

import (
	"net/http"
	"strconv"

	"corebank/backend/internal/services"
	"corebank/backend/internal/util"
)

type BeneficiaryHandler struct {
	beneficiaries *services.BeneficiaryService
}

func NewBeneficiaryHandler(b *services.BeneficiaryService) *BeneficiaryHandler {
	return &BeneficiaryHandler{beneficiaries: b}
}

func (h *BeneficiaryHandler) ListByCustomer(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "customerId")
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

type addBeneficiaryRequest struct {
	CustomerID    int64  `json:"customerId"`
	Name          string `json:"name"`
	AccountNumber string `json:"accountNumber"`
	IFSCCode      string `json:"ifscCode"`
	Nickname      string `json:"nickname"`
}

func (h *BeneficiaryHandler) Add(w http.ResponseWriter, r *http.Request) {
	var req addBeneficiaryRequest
	if err := util.DecodeJSON(r, &req); err != nil {
		util.WriteError(w, err)
		return
	}
	b, err := h.beneficiaries.Add(r.Context(), req.CustomerID, req.Name, req.AccountNumber, req.IFSCCode, req.Nickname)
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusCreated, b)
}

func (h *BeneficiaryHandler) Remove(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		util.WriteError(w, err)
		return
	}
	cid, err := strconv.ParseInt(r.URL.Query().Get("customerId"), 10, 64)
	if err != nil {
		util.WriteError(w, util.BadRequest("customerId query param is required"))
		return
	}
	if err := h.beneficiaries.Remove(r.Context(), cid, id); err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteData(w, http.StatusOK, map[string]bool{"deleted": true})
}
