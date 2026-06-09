package risk

import (
	"context"
	"sync"
	"time"

	"github.com/fintech/core/internal/config"
	"github.com/fintech/core/internal/db"
	"github.com/fintech/core/internal/model"
	"github.com/fintech/core/pkg/id"
	"github.com/fintech/core/pkg/money"
	fmterrors "github.com/fintech/core/pkg/errors"
)

type Service struct {
	db     *db.DB
	config *config.RiskConfig

	blacklistMutex sync.RWMutex
	blacklistSet   map[string]bool
}

type CheckTransferParams struct {
	FromAccountID string
	ToAccountID   string
	Amount        money.Money
	IP            string
}

func NewService(database *db.DB, cfg *config.RiskConfig) *Service {
	svc := &Service{
		db:           database,
		config:       cfg,
		blacklistSet: make(map[string]bool),
	}
	return svc
}

func (s *Service) LoadBlacklist(ctx context.Context) error {
	s.blacklistMutex.Lock()
	defer s.blacklistMutex.Unlock()

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, type, value, reason, expires_at, created_at
		FROM blacklist WHERE expires_at IS NULL OR expires_at > ?
	`, time.Now())
	if err != nil {
		return err
	}
	defer rows.Close()

	s.blacklistSet = make(map[string]bool)
	for rows.Next() {
		entry, err := db.ScanBlacklist(rows)
		if err != nil {
			return err
		}
		key := string(entry.Type) + ":" + entry.Value
		s.blacklistSet[key] = true
	}
	return nil
}

func (s *Service) IsBlacklisted(ctx context.Context, accountID string) bool {
	s.blacklistMutex.RLock()
	defer s.blacklistMutex.RUnlock()
	return s.blacklistSet[string(model.BlacklistTypeAccount)+":"+accountID]
}

func (s *Service) IsIPBlacklisted(ctx context.Context, ip string) bool {
	s.blacklistMutex.RLock()
	defer s.blacklistMutex.RUnlock()
	return s.blacklistSet[string(model.BlacklistTypeIP)+":"+ip]
}

func (s *Service) AddToBlacklist(ctx context.Context, entryType model.BlacklistType, value, reason string, expiresAt *time.Time) error {
	id := id.NewAccountID()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO blacklist (id, type, value, reason, expires_at)
		VALUES (?, ?, ?, ?, ?)
	`, id, entryType, value, reason, db.NullTime(expiresAt))
	if err != nil {
		return err
	}

	s.blacklistMutex.Lock()
	s.blacklistSet[string(entryType)+":"+value] = true
	s.blacklistMutex.Unlock()

	return nil
}

func (s *Service) RemoveFromBlacklist(ctx context.Context, entryType model.BlacklistType, value string) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM blacklist WHERE type = ? AND value = ?
	`, entryType, value)
	if err != nil {
		return err
	}

	s.blacklistMutex.Lock()
	delete(s.blacklistSet, string(entryType)+":"+value)
	s.blacklistMutex.Unlock()

	return nil
}

func (s *Service) CheckTransfer(ctx context.Context, params CheckTransferParams) error {
	if s.IsBlacklisted(ctx, params.FromAccountID) {
		return fmterrors.ErrBlacklisted
	}
	if s.IsBlacklisted(ctx, params.ToAccountID) {
		return fmterrors.ErrBlacklisted
	}
	if params.IP != "" && s.IsIPBlacklisted(ctx, params.IP) {
		return fmterrors.ErrBlacklisted
	}

	if params.Amount.GreaterThan(money.New(s.config.SingleLimit, money.DefaultCurrency)) {
		return fmterrors.ErrRiskLimitExceeded
	}

	if err := s.checkDailyLimit(ctx, params.FromAccountID, params.Amount); err != nil {
		return err
	}

	if err := s.checkMonthlyLimit(ctx, params.FromAccountID, params.Amount); err != nil {
		return err
	}

	return nil
}

func (s *Service) checkDailyLimit(ctx context.Context, accountID string, amount money.Money) error {
	today := time.Now().Format("2006-01-02")
	startOfDay := today + " 00:00:00"
	endOfDay := today + " 23:59:59"

	var totalStr string
	err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount), '0') FROM ledger_entries
		WHERE account_id = ? AND direction = 'out' AND type LIKE 'transfer%'
		AND created_at BETWEEN ? AND ?
	`, accountID, startOfDay, endOfDay).Scan(&totalStr)
	if err != nil {
		return err
	}

	total, _ := money.NewFromString(totalStr)
	newTotal := total.Add(amount)
	dailyLimit := money.New(s.config.DailyLimit, money.DefaultCurrency)

	if newTotal.GreaterThan(dailyLimit) {
		return fmterrors.ErrRiskLimitExceeded
	}
	return nil
}

func (s *Service) checkMonthlyLimit(ctx context.Context, accountID string, amount money.Money) error {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02 15:04:05")
	endOfMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, -1, now.Location()).Format("2006-01-02 15:04:05")

	var totalStr string
	err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount), '0') FROM ledger_entries
		WHERE account_id = ? AND direction = 'out' AND type LIKE 'transfer%'
		AND created_at BETWEEN ? AND ?
	`, accountID, startOfMonth, endOfMonth).Scan(&totalStr)
	if err != nil {
		return err
	}

	total, _ := money.NewFromString(totalStr)
	newTotal := total.Add(amount)
	monthlyLimit := money.New(s.config.MonthlyLimit, money.DefaultCurrency)

	if newTotal.GreaterThan(monthlyLimit) {
		return fmterrors.ErrRiskLimitExceeded
	}
	return nil
}

func (s *Service) DetectSuspicious(ctx context.Context, accountID string, amount money.Money) ([]*model.SuspiciousTransaction, error) {
	var suspicious []*model.SuspiciousTransaction
	largeAmount := money.New(s.config.LargeAmount, money.DefaultCurrency)

	if amount.GreaterThanOrEqual(largeAmount) {
		susp, err := s.createSuspicious(ctx, accountID, "", model.SuspiciousTypeLargeAmount,
			"单笔交易金额超过大额阈值", 80)
		if err != nil {
			return nil, err
		}
		suspicious = append(suspicious, susp)
	}

	windowStart := time.Now().Add(-s.config.SuspiciousWindow)
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM transfers
		WHERE from_account_id = ? AND created_at > ? AND status = 'success'
	`, accountID, windowStart).Scan(&count)
	if err != nil {
		return nil, err
	}

	if count >= s.config.SuspiciousCount {
		susp, err := s.createSuspicious(ctx, accountID, "", model.SuspiciousTypeFrequent,
			"短时间内交易频繁", 60)
		if err != nil {
			return nil, err
		}
		suspicious = append(suspicious, susp)
	}

	return suspicious, nil
}

func (s *Service) createSuspicious(ctx context.Context, accountID, transferID string, suspType model.SuspiciousType, desc string, score int) (*model.SuspiciousTransaction, error) {
	id := id.NewTransactionID()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO suspicious_transactions (id, account_id, transfer_id, type, description, risk_score)
		VALUES (?, ?, ?, ?, ?, ?)
	`, id, accountID, transferID, suspType, desc, score)
	if err != nil {
		return nil, err
	}
	return &model.SuspiciousTransaction{
		ID:          id,
		AccountID:   accountID,
		TransferID:  transferID,
		Type:        suspType,
		Description: desc,
		RiskScore:   score,
	}, nil
}

func (s *Service) ListBlacklist(ctx context.Context, limit, offset int) ([]*model.BlacklistEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, type, value, reason, expires_at, created_at
		FROM blacklist ORDER BY created_at DESC LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*model.BlacklistEntry
	for rows.Next() {
		e, err := db.ScanBlacklist(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}
