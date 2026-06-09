package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/fintech/core/internal/model"
	"github.com/fintech/core/pkg/money"
)

type createFixedDepositRequest struct {
	Principal    string `json:"principal"`
	InterestType string `json:"interest_type"`
	AutoRenew    bool   `json:"auto_renew"`
}

func (h *Handler) ListFixedDeposits(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "accountID")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	deposits, err := h.interest.ListFixedDeposits(getContext(r), accountID, limit, offset)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data":  deposits,
		"total": len(deposits),
	})
}

func (h *Handler) CreateFixedDeposit(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "accountID")

	var req createFixedDepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err)
		return
	}

	principal, err := money.NewFromString(req.Principal)
	if err != nil {
		WriteError(w, err)
		return
	}

	deposit, err := h.interest.CreateFixedDeposit(getContext(r), accountID, principal, model.InterestType(req.InterestType), req.AutoRenew)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, deposit)
}
