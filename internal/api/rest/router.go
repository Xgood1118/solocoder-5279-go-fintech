package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/fintech/core/internal/account"
	"github.com/fintech/core/internal/audit"
	"github.com/fintech/core/internal/compliance"
	"github.com/fintech/core/internal/interest"
	"github.com/fintech/core/internal/ledger"
	"github.com/fintech/core/internal/report"
	"github.com/fintech/core/internal/risk"
	"github.com/fintech/core/internal/transfer"
)

type Handler struct {
	account    *account.Service
	transfer   *transfer.Service
	ledger     *ledger.Service
	risk       *risk.Service
	audit      *audit.Service
	compliance *compliance.Service
	report     *report.Service
	interest   *interest.Service
}

func NewHandler(
	accountSvc *account.Service,
	transferSvc *transfer.Service,
	ledgerSvc *ledger.Service,
	riskSvc *risk.Service,
	auditSvc *audit.Service,
	complianceSvc *compliance.Service,
	reportSvc *report.Service,
	interestSvc *interest.Service,
) *Handler {
	return &Handler{
		account:    accountSvc,
		transfer:   transferSvc,
		ledger:     ledgerSvc,
		risk:       riskSvc,
		audit:      auditSvc,
		compliance: complianceSvc,
		report:     reportSvc,
		interest:   interestSvc,
	}
}

func (h *Handler) Router() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Biz-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", h.HealthCheck)

		r.Route("/accounts", func(r chi.Router) {
			r.Post("/", h.CreateAccount)
			r.Get("/", h.ListAccounts)
			r.Get("/{id}", h.GetAccount)
			r.Get("/{id}/balance", h.GetBalance)
			r.Post("/{id}/freeze", h.FreezeAccount)
			r.Post("/{id}/unfreeze", h.UnfreezeAccount)
			r.Post("/{id}/close", h.CloseAccount)
			r.Post("/{id}/password", h.ChangePassword)
		})

		r.Route("/transfers", func(r chi.Router) {
			r.Post("/intra-bank", h.IntraBankTransfer)
			r.Post("/cross-bank", h.CrossBankTransfer)
			r.Get("/{id}", h.GetTransfer)
			r.Get("/", h.ListTransfers)
		})

		r.Route("/ledger", func(r chi.Router) {
			r.Get("/entries", h.ListLedgerEntries)
			r.Get("/entries/{id}", h.GetLedgerEntry)
		})

		r.Route("/risk", func(r chi.Router) {
			r.Get("/blacklist", h.ListBlacklist)
			r.Post("/blacklist", h.AddToBlacklist)
			r.Delete("/blacklist", h.RemoveFromBlacklist)
			r.Get("/suspicious", h.ListSuspicious)
		})

		r.Route("/audit", func(r chi.Router) {
			r.Get("/logs", h.ListAuditLogs)
		})

		r.Route("/compliance", func(r chi.Router) {
			r.Get("/reports", h.ListComplianceReports)
			r.Post("/process", h.ProcessComplianceReports)
		})

		r.Route("/reports", func(r chi.Router) {
			r.Get("/daily/{accountID}", h.GetDailyReport)
			r.Post("/daily/{accountID}", h.GenerateDailyReport)
			r.Get("/monthly/{accountID}", h.GetMonthlyReport)
			r.Post("/monthly/{accountID}", h.GenerateMonthlyReport)
			r.Get("/reconciliation", h.ListReconciliationDiffs)
			r.Post("/reconciliation", h.RunReconciliation)
		})

		r.Route("/interest", func(r chi.Router) {
			r.Get("/fixed/{accountID}", h.ListFixedDeposits)
			r.Post("/fixed/{accountID}", h.CreateFixedDeposit)
		})
	})

	return r
}

func (h *Handler) getIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	return ip
}

func (h *Handler) getBizID(r *http.Request) string {
	return r.Header.Get("X-Biz-ID")
}

func (h *Handler) getRequestID(r *http.Request) string {
	return middleware.GetReqID(r.Context())
}

func getContext(r *http.Request) context.Context {
	return r.Context()
}
