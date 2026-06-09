package rest

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListLedgerEntries(w http.ResponseWriter, r *http.Request) {
	accountID := r.URL.Query().Get("account_id")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if accountID == "" {
		WriteError(w, nil)
		return
	}

	entries, err := h.ledger.ListByAccount(getContext(r), accountID, limit, offset)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data":  entries,
		"total": len(entries),
	})
}

func (h *Handler) GetLedgerEntry(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	entry, err := h.ledger.GetEntry(getContext(r), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, entry)
}
