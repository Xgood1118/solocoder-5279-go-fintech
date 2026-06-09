package rest

import (
	"encoding/json"
	"net/http"

	fmterrors "github.com/fintech/core/pkg/errors"
)

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, err error) {
	if err == nil {
		problem := fmterrors.NewInternal("unknown error")
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(problem)
		return
	}
	if problem, ok := err.(*fmterrors.Problem); ok {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(problem.Status)
		_ = json.NewEncoder(w).Encode(problem)
		return
	}

	problem := fmterrors.NewInternal(err.Error())
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(problem)
}

func WriteProblem(w http.ResponseWriter, problem *fmterrors.Problem) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(problem.Status)
	_ = json.NewEncoder(w).Encode(problem)
}
