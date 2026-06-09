package compliance

import (
	"context"
	"log"
	"time"

	"github.com/fintech/core/internal/config"
	"github.com/fintech/core/internal/db"
	"github.com/fintech/core/internal/model"
	"github.com/fintech/core/pkg/id"
	"github.com/fintech/core/pkg/money"
)

type Service struct {
	db     *db.DB
	config *config.RiskConfig
}

func NewService(database *db.DB, cfg *config.RiskConfig) *Service {
	return &Service{
		db:     database,
		config: cfg,
	}
}

func (s *Service) CheckAndReportLargeAmount(ctx context.Context, accountID, transferID string, amount money.Money) error {
	largeAmount := money.New(s.config.LargeAmount, money.DefaultCurrency)
	if amount.LessThan(largeAmount) {
		return nil
	}

	content := "大额交易上报：账户 " + accountID + "，交易 " + transferID + "，金额 " + amount.String()
	_, err := s.createReport(ctx, model.ReportTypeLargeAmount, accountID, transferID, content)
	return err
}

func (s *Service) CheckAndReportSuspicious(ctx context.Context, accountID, transferID string, suspiciousCount int) error {
	if suspiciousCount < s.config.SuspiciousCount {
		return nil
	}

	content := "可疑交易上报：账户 " + accountID + "，当日可疑交易 " + string(rune(suspiciousCount)) + " 笔"
	_, err := s.createReport(ctx, model.ReportTypeSuspicious, accountID, transferID, content)
	return err
}

func (s *Service) createReport(ctx context.Context, reportType model.ReportType, accountID, transferID, content string) (*model.ComplianceReport, error) {
	id := id.NewReportID()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO compliance_reports (id, report_type, account_id, transfer_id, report_content, status)
		VALUES (?, ?, ?, ?, ?, 'pending')
	`, id, reportType, accountID, transferID, content)
	if err != nil {
		return nil, err
	}
	return &model.ComplianceReport{
		ID:            id,
		ReportType:    reportType,
		AccountID:     accountID,
		TransferID:    transferID,
		ReportContent: content,
		Status:        model.ReportStatusPending,
	}, nil
}

func (s *Service) ProcessPendingReports(ctx context.Context) (int, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, report_type, account_id, transfer_id, report_content, status, retry_count, report_batch_no, reported_at, failure_reason, created_at
		FROM compliance_reports 
		WHERE status = 'pending' OR (status = 'failed' AND retry_count < 3)
		LIMIT 100
	`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var reports []*model.ComplianceReport
	for rows.Next() {
		r, err := db.ScanComplianceReport(rows)
		if err != nil {
			return 0, err
		}
		reports = append(reports, r)
	}

	successCount := 0
	for _, r := range reports {
		if err := s.reportToCentralBank(ctx, r); err != nil {
			log.Printf("Report %s failed: %v", r.ID, err)
		} else {
			successCount++
		}
	}

	return successCount, nil
}

func (s *Service) reportToCentralBank(ctx context.Context, report *model.ComplianceReport) error {
	success := mockCentralBankReport(report)

	if success {
		now := time.Now()
		_, err := s.db.ExecContext(ctx, `
			UPDATE compliance_reports SET status = 'reported', reported_at = ?, retry_count = retry_count + 1 WHERE id = ?
		`, now, report.ID)
		return err
	}

	_, err := s.db.ExecContext(ctx, `
		UPDATE compliance_reports SET status = 'failed', failure_reason = '央行接口调用失败', retry_count = retry_count + 1 WHERE id = ?
	`, report.ID)
	return err
}

func mockCentralBankReport(report *model.ComplianceReport) bool {
	log.Printf("Mock report to central bank: %s, type: %s", report.ID, report.ReportType)
	time.Sleep(50 * time.Millisecond)
	return true
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]*model.ComplianceReport, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, report_type, account_id, transfer_id, report_content, status, retry_count, report_batch_no, reported_at, failure_reason, created_at
		FROM compliance_reports ORDER BY created_at DESC LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*model.ComplianceReport
	for rows.Next() {
		r, err := db.ScanComplianceReport(rows)
		if err != nil {
			return nil, err
		}
		reports = append(reports, r)
	}
	return reports, nil
}
