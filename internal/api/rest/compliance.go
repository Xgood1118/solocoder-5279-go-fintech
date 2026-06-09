package rest

import (
	"net/http"
	"strconv"
)

func (h *Handler) ListComplianceReports(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	reports, err := h.compliance.List(getContext(r), limit, offset)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data":  reports,
		"total": len(reports),
	})
}

func (h *Handler) ProcessComplianceReports(w http.ResponseWriter, r *http.Request) {
	count, err := h.compliance.ProcessPendingReports(getContext(r))
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"processed": count,
	})
}
