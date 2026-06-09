package rest

import "net/http"

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "fintech-core",
	})
}
