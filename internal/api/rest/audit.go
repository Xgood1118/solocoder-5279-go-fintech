package rest

import (
	"net/http"
	"strconv"
)

func (h *Handler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	accountID := r.URL.Query().Get("account_id")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	var logs interface{}
	var err error

	if accountID != "" {
		logs, err = h.audit.ListByAccount(getContext(r), accountID, limit, offset)
	} else {
		logs, err = h.audit.List(getContext(r), limit, offset)
	}

	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data":  logs,
		"total": 0,
	})
}
