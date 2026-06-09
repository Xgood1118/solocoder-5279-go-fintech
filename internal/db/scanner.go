package db

import (
	"database/sql"
	"time"

	"github.com/fintech/core/internal/model"
	"github.com/fintech/core/pkg/money"
	"github.com/shopspring/decimal"
)

func ScanAccount(rows *sql.Rows) (*model.Account, error) {
	var a model.Account
	var balanceStr, frozenStr string
	var closedAt sql.NullTime
	err := rows.Scan(
		&a.ID, &a.AccountType, &a.AccountNo, &a.Name, &a.IDCardNo,
		&a.PasswordHash, &a.Status, &balanceStr, &frozenStr,
		&a.CreatedAt, &a.UpdatedAt, &closedAt,
	)
	if err != nil {
		return nil, err
	}
	bal, _ := decimal.NewFromString(balanceStr)
	a.Balance = money.New(bal, money.DefaultCurrency)
	frozen, _ := decimal.NewFromString(frozenStr)
	a.FrozenAmount = money.New(frozen, money.DefaultCurrency)
	if closedAt.Valid {
		t := closedAt.Time
		a.ClosedAt = &t
	}
	return &a, nil
}

func ScanLedgerEntry(rows *sql.Rows) (*model.LedgerEntry, error) {
	var e model.LedgerEntry
	var amountStr, balanceStr string
	err := rows.Scan(
		&e.ID, &e.AccountID, &e.TransferID, &e.BizID,
		&e.Direction, &e.Type, &amountStr, &balanceStr,
		&e.Counterparty, &e.Remark, &e.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	amt, _ := decimal.NewFromString(amountStr)
	e.Amount = money.New(amt, money.DefaultCurrency)
	bal, _ := decimal.NewFromString(balanceStr)
	e.BalanceAfter = money.New(bal, money.DefaultCurrency)
	return &e, nil
}

func ScanTransfer(rows *sql.Rows) (*model.Transfer, error) {
	var t model.Transfer
	var amountStr, feeStr string
	var toAccountID, toBankCode, toAccountNo, toAccountName sql.NullString
	var settlementID, failureReason sql.NullString
	var completedAt sql.NullTime
	err := rows.Scan(
		&t.ID, &t.BizID, &t.FromAccountID, &toAccountID, &toBankCode,
		&toAccountNo, &toAccountName, &amountStr, &feeStr, &t.Remark,
		&t.Status, &t.IsCrossBank, &settlementID, &failureReason,
		&t.CreatedAt, &t.UpdatedAt, &completedAt,
	)
	if err != nil {
		return nil, err
	}
	amt, _ := decimal.NewFromString(amountStr)
	t.Amount = money.New(amt, money.DefaultCurrency)
	f, _ := decimal.NewFromString(feeStr)
	t.Fee = money.New(f, money.DefaultCurrency)
	t.ToAccountID = toAccountID.String
	t.ToBankCode = toBankCode.String
	t.ToAccountNo = toAccountNo.String
	t.ToAccountName = toAccountName.String
	t.SettlementID = settlementID.String
	t.FailureReason = failureReason.String
	if completedAt.Valid {
		ct := completedAt.Time
		t.CompletedAt = &ct
	}
	return &t, nil
}

func ScanAuditLog(rows *sql.Rows) (*model.AuditLog, error) {
	var l model.AuditLog
	var accountID, operatorID, ip, userAgent, requestID sql.NullString
	var before, after, detail sql.NullString
	err := rows.Scan(
		&l.ID, &l.Action, &accountID, &operatorID, &ip,
		&userAgent, &requestID, &before, &after, &detail, &l.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	l.AccountID = accountID.String
	l.OperatorID = operatorID.String
	l.IP = ip.String
	l.UserAgent = userAgent.String
	l.RequestID = requestID.String
	l.Before = before.String
	l.After = after.String
	l.Detail = detail.String
	return &l, nil
}

func ScanBlacklist(rows *sql.Rows) (*model.BlacklistEntry, error) {
	var e model.BlacklistEntry
	var reason sql.NullString
	var expiresAt sql.NullTime
	err := rows.Scan(&e.ID, &e.Type, &e.Value, &reason, &expiresAt, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	e.Reason = reason.String
	if expiresAt.Valid {
		t := expiresAt.Time
		e.ExpiresAt = &t
	}
	return &e, nil
}

func ScanSuspicious(rows *sql.Rows) (*model.SuspiciousTransaction, error) {
	var s model.SuspiciousTransaction
	var desc sql.NullString
	var reportedAt sql.NullTime
	err := rows.Scan(
		&s.ID, &s.AccountID, &s.TransferID, &s.Type, &desc,
		&s.RiskScore, &s.Reported, &reportedAt, &s.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	s.Description = desc.String
	if reportedAt.Valid {
		t := reportedAt.Time
		s.ReportedAt = &t
	}
	return &s, nil
}

func ScanComplianceReport(rows *sql.Rows) (*model.ComplianceReport, error) {
	var r model.ComplianceReport
	var accountID, transferID, content, batchNo, failureReason sql.NullString
	var reportedAt sql.NullTime
	err := rows.Scan(
		&r.ID, &r.ReportType, &accountID, &transferID, &content,
		&r.Status, &r.RetryCount, &batchNo, &reportedAt,
		&failureReason, &r.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	r.AccountID = accountID.String
	r.TransferID = transferID.String
	r.ReportContent = content.String
	r.ReportBatchNo = batchNo.String
	r.FailureReason = failureReason.String
	if reportedAt.Valid {
		t := reportedAt.Time
		r.ReportedAt = &t
	}
	return &r, nil
}

func ScanDailyReport(rows *sql.Rows) (*model.DailyAccountReport, error) {
	var r model.DailyAccountReport
	var beginStr, endStr, incomeStr, expenseStr string
	err := rows.Scan(
		&r.ID, &r.AccountID, &r.ReportDate, &beginStr, &endStr,
		&incomeStr, &expenseStr, &r.TxnCount, &r.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	begin, _ := decimal.NewFromString(beginStr)
	r.BeginBalance = money.New(begin, money.DefaultCurrency)
	end, _ := decimal.NewFromString(endStr)
	r.EndBalance = money.New(end, money.DefaultCurrency)
	income, _ := decimal.NewFromString(incomeStr)
	r.TotalIncome = money.New(income, money.DefaultCurrency)
	expense, _ := decimal.NewFromString(expenseStr)
	r.TotalExpense = money.New(expense, money.DefaultCurrency)
	return &r, nil
}

func ScanMonthlyReport(rows *sql.Rows) (*model.MonthlyAccountReport, error) {
	var r model.MonthlyAccountReport
	var beginStr, endStr, incomeStr, expenseStr, interestStr string
	err := rows.Scan(
		&r.ID, &r.AccountID, &r.ReportMonth, &beginStr, &endStr,
		&incomeStr, &expenseStr, &interestStr, &r.TxnCount, &r.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	begin, _ := decimal.NewFromString(beginStr)
	r.BeginBalance = money.New(begin, money.DefaultCurrency)
	end, _ := decimal.NewFromString(endStr)
	r.EndBalance = money.New(end, money.DefaultCurrency)
	income, _ := decimal.NewFromString(incomeStr)
	r.TotalIncome = money.New(income, money.DefaultCurrency)
	expense, _ := decimal.NewFromString(expenseStr)
	r.TotalExpense = money.New(expense, money.DefaultCurrency)
	interest, _ := decimal.NewFromString(interestStr)
	r.Interest = money.New(interest, money.DefaultCurrency)
	return &r, nil
}

func ScanReconDiff(rows *sql.Rows) (*model.ReconciliationDiff, error) {
	var r model.ReconciliationDiff
	var balStr, ledgerStr, diffStr string
	var fixedAt sql.NullTime
	err := rows.Scan(
		&r.ID, &r.AccountID, &r.CheckDate, &balStr, &ledgerStr, &diffStr,
		&r.IsFixed, &fixedAt, &r.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	bal, _ := decimal.NewFromString(balStr)
	r.BalanceAmount = money.New(bal, money.DefaultCurrency)
	ledger, _ := decimal.NewFromString(ledgerStr)
	r.LedgerAmount = money.New(ledger, money.DefaultCurrency)
	diff, _ := decimal.NewFromString(diffStr)
	r.DiffAmount = money.New(diff, money.DefaultCurrency)
	if fixedAt.Valid {
		t := fixedAt.Time
		r.FixedAt = &t
	}
	return &r, nil
}

func ScanFixedDeposit(rows *sql.Rows) (*model.FixedDeposit, error) {
	var f model.FixedDeposit
	var principalStr, interestStr string
	var redeemedAt sql.NullTime
	err := rows.Scan(
		&f.ID, &f.AccountID, &principalStr, &f.InterestType, &f.AnnualRate,
		&interestStr, &f.StartDate, &f.MaturityDate, &f.AutoRenew,
		&f.IsRedeemed, &redeemedAt, &f.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	p, _ := decimal.NewFromString(principalStr)
	f.Principal = money.New(p, money.DefaultCurrency)
	i, _ := decimal.NewFromString(interestStr)
	f.InterestAmount = money.New(i, money.DefaultCurrency)
	if redeemedAt.Valid {
		t := redeemedAt.Time
		f.RedeemedAt = &t
	}
	return &f, nil
}

func ScanInterestRecord(rows *sql.Rows) (*model.InterestRecord, error) {
	var r model.InterestRecord
	var principalStr, interestStr string
	err := rows.Scan(
		&r.ID, &r.AccountID, &r.LedgerID, &r.InterestType, &principalStr,
		&r.Rate, &r.Days, &interestStr, &r.PeriodStart, &r.PeriodEnd, &r.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	p, _ := decimal.NewFromString(principalStr)
	r.Principal = money.New(p, money.DefaultCurrency)
	i, _ := decimal.NewFromString(interestStr)
	r.Interest = money.New(i, money.DefaultCurrency)
	return &r, nil
}

func ScanInterestAccrual(rows *sql.Rows) (*model.InterestAccrual, error) {
	var a model.InterestAccrual
	var principalStr, accruedStr string
	err := rows.Scan(&a.ID, &a.AccountID, &a.AccrualDate, &principalStr, &accruedStr, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	p, _ := decimal.NewFromString(principalStr)
	a.Principal = money.New(p, money.DefaultCurrency)
	acc, _ := decimal.NewFromString(accruedStr)
	a.Accrued = money.New(acc, money.DefaultCurrency)
	return &a, nil
}

func NullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func NullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
