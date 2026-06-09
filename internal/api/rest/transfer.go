package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/fintech/core/internal/audit"
	"github.com/fintech/core/internal/model"
	"github.com/fintech/core/internal/transfer"
	fmterrors "github.com/fintech/core/pkg/errors"
	"github.com/fintech/core/pkg/money"
)

type intraBankTransferRequest struct {
	FromAccountID string `json:"from_account_id"`
	ToAccountID   string `json:"to_account_id"`
	Amount        string `json:"amount"`
	Remark        string `json:"remark"`
	Password      string `json:"password"`
}

func (h *Handler) IntraBankTransfer(w http.ResponseWriter, r *http.Request) {
	var req intraBankTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err)
		return
	}

	amount, err := money.NewFromString(req.Amount)
	if err != nil {
		WriteError(w, err)
		return
	}

	bizID := h.getBizID(r)
	if bizID == "" {
		bizID = "rest-" + h.getRequestID(r)
	}

	params := transfer.IntraBankTransferParams{
		BizID:         bizID,
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        amount,
		Remark:        req.Remark,
		Password:      req.Password,
		IP:            h.getIP(r),
	}

	t, err := h.transfer.IntraBankTransfer(getContext(r), params)
	if err != nil {
		WriteError(w, err)
		return
	}

	h.audit.Log(getContext(r), audit.LogParams{
		Action:    model.AuditActionTransfer,
		AccountID: req.FromAccountID,
		IP:        h.getIP(r),
		RequestID: h.getRequestID(r),
		Detail:    "行内转账: " + req.Amount,
	})

	WriteJSON(w, http.StatusCreated, t)
}

type crossBankTransferRequest struct {
	FromAccountID string `json:"from_account_id"`
	ToBankCode    string `json:"to_bank_code"`
	ToAccountNo   string `json:"to_account_no"`
	ToAccountName string `json:"to_account_name"`
	Amount        string `json:"amount"`
	Remark        string `json:"remark"`
	Password      string `json:"password"`
}

func (h *Handler) CrossBankTransfer(w http.ResponseWriter, r *http.Request) {
	var req crossBankTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err)
		return
	}

	amount, err := money.NewFromString(req.Amount)
	if err != nil {
		WriteError(w, err)
		return
	}

	bizID := h.getBizID(r)
	if bizID == "" {
		bizID = "rest-cross-" + h.getRequestID(r)
	}

	params := transfer.CrossBankTransferParams{
		BizID:         bizID,
		FromAccountID: req.FromAccountID,
		ToBankCode:    req.ToBankCode,
		ToAccountNo:   req.ToAccountNo,
		ToAccountName: req.ToAccountName,
		Amount:        amount,
		Remark:        req.Remark,
		Password:      req.Password,
		IP:            h.getIP(r),
	}

	t, err := h.transfer.CrossBankTransfer(getContext(r), params)
	if err != nil {
		WriteError(w, err)
		return
	}

	h.audit.Log(getContext(r), audit.LogParams{
		Action:    model.AuditActionCrossBankTransfer,
		AccountID: req.FromAccountID,
		IP:        h.getIP(r),
		RequestID: h.getRequestID(r),
		Detail:    "跨行转账: " + req.Amount,
	})

	WriteJSON(w, http.StatusAccepted, t)
}

func (h *Handler) GetTransfer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	t, err := h.transfer.Get(getContext(r), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, t)
}

func (h *Handler) ListTransfers(w http.ResponseWriter, r *http.Request) {
	accountID := r.URL.Query().Get("account_id")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if accountID == "" {
		WriteError(w, fmterrors.NewBadRequest("MISSING_ACCOUNT_ID", "缺少 account_id 参数"))
		return
	}

	transfers, err := h.transfer.ListByAccount(getContext(r), accountID, limit, offset)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data":  transfers,
		"total": len(transfers),
	})
}
