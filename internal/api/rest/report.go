package rest

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetDailyReport(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "accountID")
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	report, err := h.report.GetDailyReport(getContext(r), accountID, date)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, report)
}

func (h *Handler) GenerateDailyReport(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "accountID")
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	report, err := h.report.GenerateDailyReport(getContext(r), accountID, date)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, report)
}

func (h *Handler) GetMonthlyReport(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "accountID")
	month := r.URL.Query().Get("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	report, err := h.report.GetMonthlyReport(getContext(r), accountID, month)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, report)
}

func (h *Handler) GenerateMonthlyReport(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "accountID")
	month := r.URL.Query().Get("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	report, err := h.report.GenerateMonthlyReport(getContext(r), accountID, month)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, report)
}

func (h *Handler) ListReconciliationDiffs(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	diffs, err := h.report.ListReconciliationDiffs(getContext(r), date, limit, offset)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data":  diffs,
		"total": len(diffs),
	})
}

func (h *Handler) RunReconciliation(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	diffs, err := h.report.ReconcileAll(getContext(r), date)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"diff_count": len(diffs),
		"diffs":      diffs,
	})
}
