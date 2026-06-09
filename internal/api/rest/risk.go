package rest

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/fintech/core/internal/model"
)

type addBlacklistRequest struct {
	Type      string  `json:"type"`
	Value     string  `json:"value"`
	Reason    string  `json:"reason"`
	ExpiresAt *string `json:"expires_at"`
}

func (h *Handler) ListBlacklist(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	entries, err := h.risk.ListBlacklist(getContext(r), limit, offset)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data":  entries,
		"total": len(entries),
	})
}

func (h *Handler) AddToBlacklist(w http.ResponseWriter, r *http.Request) {
	var req addBlacklistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err)
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err == nil {
			expiresAt = &t
		}
	}

	err := h.risk.AddToBlacklist(getContext(r), model.BlacklistType(req.Type), req.Value, req.Reason, expiresAt)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, map[string]string{
		"message": "added to blacklist",
	})
}

type removeBlacklistRequest struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (h *Handler) RemoveFromBlacklist(w http.ResponseWriter, r *http.Request) {
	var req removeBlacklistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err)
		return
	}

	err := h.risk.RemoveFromBlacklist(getContext(r), model.BlacklistType(req.Type), req.Value)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"message": "removed from blacklist",
	})
}

func (h *Handler) ListSuspicious(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data":  []interface{}{},
		"total": 0,
	})
}
