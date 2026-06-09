package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fintech/core/internal/account"
	"github.com/fintech/core/internal/api/rest"
	"github.com/fintech/core/internal/audit"
	"github.com/fintech/core/internal/compliance"
	"github.com/fintech/core/internal/config"
	"github.com/fintech/core/internal/db"
	grpcserver "github.com/fintech/core/internal/grpc"
	"github.com/fintech/core/internal/interest"
	"github.com/fintech/core/internal/ledger"
	"github.com/fintech/core/internal/report"
	"github.com/fintech/core/internal/risk"
	"github.com/fintech/core/internal/transfer"
)

func main() {
	cfg := config.Default()

	database, err := db.New(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()
	log.Println("Database connected successfully")

	ledgerSvc := ledger.NewService(database)
	accountSvc := account.NewService(database, ledgerSvc, cfg.Security.PasswordHashCost)
	riskSvc := risk.NewService(database, &cfg.Risk)
	transferSvc := transfer.NewService(database, accountSvc, ledgerSvc, riskSvc)
	auditSvc := audit.NewService(database, 1000)
	complianceSvc := compliance.NewService(database, &cfg.Risk)
	reportSvc := report.NewService(database, ledgerSvc)
	interestSvc := interest.NewService(database, ledgerSvc, &cfg.Interest)

	if err := riskSvc.LoadBlacklist(context.Background()); err != nil {
		log.Printf("Warning: failed to load blacklist: %v", err)
	}

	handler := rest.NewHandler(
		accountSvc,
		transferSvc,
		ledgerSvc,
		riskSvc,
		auditSvc,
		complianceSvc,
		reportSvc,
		interestSvc,
	)

	router := handler.Router()

	addr := fmt.Sprintf(":%d", cfg.Server.RESTPort)
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		log.Printf("REST API server starting on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	grpcAddr := fmt.Sprintf(":%d", cfg.Server.GRPCPort)
	grpcSrv := grpcserver.NewServer(grpcAddr, accountSvc, transferSvc, ledgerSvc)
	go func() {
		log.Printf("gRPC server starting on %s", grpcAddr)
		if err := grpcSrv.Start(); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	grpcSrv.Stop()
	auditSvc.Stop()
	log.Println("Server exited properly")
}
