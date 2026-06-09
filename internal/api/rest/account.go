package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/fintech/core/internal/account"
	"github.com/fintech/core/internal/audit"
	"github.com/fintech/core/internal/model"
	"github.com/fintech/core/pkg/money"
)

type createAccountRequest struct {
	AccountType    string `json:"account_type"`
	Name           string `json:"name"`
	IDCardNo       string `json:"id_card_no"`
	Password       string `json:"password"`
	InitialBalance string `json:"initial_balance"`
}

func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req createAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err)
		return
	}

	initialBalance := money.Zero()
	if req.InitialBalance != "" {
		var err error
		initialBalance, err = money.NewFromString(req.InitialBalance)
		if err != nil {
			WriteError(w, err)
			return
		}
	}

	params := account.CreateAccountParams{
		AccountType:    model.AccountType(req.AccountType),
		Name:           req.Name,
		IDCardNo:       req.IDCardNo,
		Password:       req.Password,
		InitialBalance: initialBalance,
	}

	acc, err := h.account.Create(getContext(r), params)
	if err != nil {
		WriteError(w, err)
		return
	}

	h.audit.Log(getContext(r), audit.LogParams{
		Action:    model.AuditActionCreateAccount,
		AccountID: acc.ID,
		IP:        h.getIP(r),
		RequestID: h.getRequestID(r),
		Detail:    "开户: " + req.Name,
	})

	WriteJSON(w, http.StatusCreated, acc)
}

func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	acc, err := h.account.Get(getContext(r), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, acc)
}

func (h *Handler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	accounts, err := h.account.List(getContext(r), limit, offset)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data":  accounts,
		"total": len(accounts),
	})
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	balance, err := h.account.GetBalance(getContext(r), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"account_id": id,
		"balance":    balance,
	})
}

func (h *Handler) FreezeAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.account.Freeze(getContext(r), id); err != nil {
		WriteError(w, err)
		return
	}

	h.audit.Log(getContext(r), audit.LogParams{
		Action:    model.AuditActionFreezeAccount,
		AccountID: id,
		IP:        h.getIP(r),
		RequestID: h.getRequestID(r),
		Detail:    "冻结账户",
	})

	WriteJSON(w, http.StatusOK, map[string]string{
		"status": "frozen",
	})
}

func (h *Handler) UnfreezeAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.account.Unfreeze(getContext(r), id); err != nil {
		WriteError(w, err)
		return
	}

	h.audit.Log(getContext(r), audit.LogParams{
		Action:    model.AuditActionUnfreezeAccount,
		AccountID: id,
		IP:        h.getIP(r),
		RequestID: h.getRequestID(r),
		Detail:    "解冻账户",
	})

	WriteJSON(w, http.StatusOK, map[string]string{
		"status": "active",
	})
}

func (h *Handler) CloseAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.account.Close(getContext(r), id); err != nil {
		WriteError(w, err)
		return
	}

	h.audit.Log(getContext(r), audit.LogParams{
		Action:    model.AuditActionCloseAccount,
		AccountID: id,
		IP:        h.getIP(r),
		RequestID: h.getRequestID(r),
		Detail:    "销户",
	})

	WriteJSON(w, http.StatusOK, map[string]string{
		"status": "closed",
	})
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err)
		return
	}

	if err := h.account.ChangePassword(getContext(r), id, req.OldPassword, req.NewPassword); err != nil {
		WriteError(w, err)
		return
	}

	h.audit.Log(getContext(r), audit.LogParams{
		Action:    model.AuditActionChangePassword,
		AccountID: id,
		IP:        h.getIP(r),
		RequestID: h.getRequestID(r),
		Detail:    "修改密码",
	})

	WriteJSON(w, http.StatusOK, map[string]string{
		"message": "password changed",
	})
}
