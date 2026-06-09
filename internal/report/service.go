package report

import (
	"context"
	"database/sql"
	"time"

	"github.com/fintech/core/internal/db"
	"github.com/fintech/core/internal/ledger"
	"github.com/fintech/core/internal/model"
	"github.com/fintech/core/pkg/id"
	"github.com/fintech/core/pkg/money"
)

type Service struct {
	db     *db.DB
	ledger *ledger.Service
}

func NewService(database *db.DB, ledgerSvc *ledger.Service) *Service {
	return &Service{
		db:     database,
		ledger: ledgerSvc,
	}
}

func (s *Service) GenerateDailyReport(ctx context.Context, accountID string, reportDate string) (*model.DailyAccountReport, error) {
	existing, err := s.GetDailyReport(ctx, accountID, reportDate)
	if err == nil && existing != nil {
		return existing, nil
	}

	startOfDay := reportDate + " 00:00:00"
	endOfDay := reportDate + " 23:59:59"

	var beginBalanceStr string
	err = s.db.QueryRowContext(ctx, `
		SELECT COALESCE(
			(SELECT balance_after FROM ledger_entries 
			 WHERE account_id = ? AND created_at < ? 
			 ORDER BY created_at DESC LIMIT 1),
			'0'
		)
	`, accountID, startOfDay).Scan(&beginBalanceStr)
	if err != nil {
		return nil, err
	}
	beginBalance, _ := money.NewFromString(beginBalanceStr)

	var totalIncomeStr, totalExpenseStr string
	var txnCount int
	err = s.db.QueryRowContext(ctx, `
		SELECT 
			COALESCE(SUM(CASE WHEN direction = 'in' THEN amount ELSE 0 END), '0'),
			COALESCE(SUM(CASE WHEN direction = 'out' THEN amount ELSE 0 END), '0'),
			COUNT(*)
		FROM ledger_entries 
		WHERE account_id = ? AND created_at BETWEEN ? AND ?
	`, accountID, startOfDay, endOfDay).Scan(&totalIncomeStr, &totalExpenseStr, &txnCount)
	if err != nil {
		return nil, err
	}
	totalIncome, _ := money.NewFromString(totalIncomeStr)
	totalExpense, _ := money.NewFromString(totalExpenseStr)

	endBalance := beginBalance.Add(totalIncome).Sub(totalExpense)

	reportID := id.NewReportID()
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO daily_account_reports (id, account_id, report_date, begin_balance, end_balance, total_income, total_expense, txn_count)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, reportID, accountID, reportDate, beginBalance.Amount.String(), endBalance.Amount.String(),
		totalIncome.Amount.String(), totalExpense.Amount.String(), txnCount)
	if err != nil {
		return nil, err
	}

	return &model.DailyAccountReport{
		ID:           reportID,
		AccountID:    accountID,
		ReportDate:   reportDate,
		BeginBalance: beginBalance,
		EndBalance:   endBalance,
		TotalIncome:  totalIncome,
		TotalExpense: totalExpense,
		TxnCount:     txnCount,
	}, nil
}

func (s *Service) GetDailyReport(ctx context.Context, accountID, reportDate string) (*model.DailyAccountReport, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, account_id, report_date, begin_balance, end_balance, total_income, total_expense, txn_count, created_at
		FROM daily_account_reports WHERE account_id = ? AND report_date = ?
	`, accountID, reportDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	return db.ScanDailyReport(rows)
}

func (s *Service) GenerateMonthlyReport(ctx context.Context, accountID string, reportMonth string) (*model.MonthlyAccountReport, error) {
	existing, err := s.GetMonthlyReport(ctx, accountID, reportMonth)
	if err == nil && existing != nil {
		return existing, nil
	}

	startOfMonth := reportMonth + "-01 00:00:00"
	t, _ := time.Parse("2006-01", reportMonth)
	endOfMonth := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, -1, t.Location()).Format("2006-01-02 15:04:05")

	var beginBalanceStr string
	err = s.db.QueryRowContext(ctx, `
		SELECT COALESCE(
			(SELECT balance_after FROM ledger_entries 
			 WHERE account_id = ? AND created_at < ? 
			 ORDER BY created_at DESC LIMIT 1),
			'0'
		)
	`, accountID, startOfMonth).Scan(&beginBalanceStr)
	if err != nil {
		return nil, err
	}
	beginBalance, _ := money.NewFromString(beginBalanceStr)

	var totalIncomeStr, totalExpenseStr, interestStr string
	var txnCount int
	err = s.db.QueryRowContext(ctx, `
		SELECT 
			COALESCE(SUM(CASE WHEN direction = 'in' AND type != 'interest' THEN amount ELSE 0 END), '0'),
			COALESCE(SUM(CASE WHEN direction = 'out' THEN amount ELSE 0 END), '0'),
			COALESCE(SUM(CASE WHEN type = 'interest' THEN amount ELSE 0 END), '0'),
			COUNT(*)
		FROM ledger_entries 
		WHERE account_id = ? AND created_at BETWEEN ? AND ?
	`, accountID, startOfMonth, endOfMonth).Scan(&totalIncomeStr, &totalExpenseStr, &interestStr, &txnCount)
	if err != nil {
		return nil, err
	}
	totalIncome, _ := money.NewFromString(totalIncomeStr)
	totalExpense, _ := money.NewFromString(totalExpenseStr)
	interest, _ := money.NewFromString(interestStr)

	endBalance := beginBalance.Add(totalIncome).Add(interest).Sub(totalExpense)

	reportID := id.NewReportID()
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO monthly_account_reports (id, account_id, report_month, begin_balance, end_balance, total_income, total_expense, interest, txn_count)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, reportID, accountID, reportMonth, beginBalance.Amount.String(), endBalance.Amount.String(),
		totalIncome.Amount.String(), totalExpense.Amount.String(), interest.Amount.String(), txnCount)
	if err != nil {
		return nil, err
	}

	return &model.MonthlyAccountReport{
		ID:           reportID,
		AccountID:    accountID,
		ReportMonth:  reportMonth,
		BeginBalance: beginBalance,
		EndBalance:   endBalance,
		TotalIncome:  totalIncome,
		TotalExpense: totalExpense,
		Interest:     interest,
		TxnCount:     txnCount,
	}, nil
}

func (s *Service) GetMonthlyReport(ctx context.Context, accountID, reportMonth string) (*model.MonthlyAccountReport, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, account_id, report_month, begin_balance, end_balance, total_income, total_expense, interest, txn_count, created_at
		FROM monthly_account_reports WHERE account_id = ? AND report_month = ?
	`, accountID, reportMonth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	return db.ScanMonthlyReport(rows)
}

func (s *Service) ReconcileAll(ctx context.Context, checkDate string) ([]*model.ReconciliationDiff, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id FROM accounts WHERE status != 'closed'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accountIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		accountIDs = append(accountIDs, id)
	}

	var diffs []*model.ReconciliationDiff
	for _, accountID := range accountIDs {
		diff, err := s.ledger.ReconcileAccount(ctx, accountID, checkDate)
		if err != nil {
			return nil, err
		}
		if diff != nil {
			diffs = append(diffs, diff)
		}
	}

	return diffs, nil
}

func (s *Service) ListReconciliationDiffs(ctx context.Context, checkDate string, limit, offset int) ([]*model.ReconciliationDiff, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, account_id, check_date, balance_amount, ledger_amount, diff_amount, is_fixed, fixed_at, created_at
		FROM reconciliation_diffs WHERE check_date = ? ORDER BY created_at DESC LIMIT ? OFFSET ?
	`, checkDate, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var diffs []*model.ReconciliationDiff
	for rows.Next() {
		d, err := db.ScanReconDiff(rows)
		if err != nil {
			return nil, err
		}
		diffs = append(diffs, d)
	}
	return diffs, nil
}
