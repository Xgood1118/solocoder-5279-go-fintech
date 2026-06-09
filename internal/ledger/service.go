package ledger

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/fintech/core/internal/db"
	"github.com/fintech/core/internal/model"
	"github.com/fintech/core/pkg/id"
	"github.com/fintech/core/pkg/money"
)

type Service struct {
	db *db.DB
}

func NewService(database *db.DB) *Service {
	return &Service{db: database}
}

type CreateEntryParams struct {
	AccountID    string
	TransferID   string
	BizID        string
	Direction    model.TransactionDirection
	Type         model.TransactionType
	Amount       money.Money
	Counterparty string
	Remark       string
}

func (s *Service) CreateEntry(ctx context.Context, tx *sql.Tx, params CreateEntryParams) (*model.LedgerEntry, error) {
	var currentBalanceStr string
	err := tx.QueryRowContext(ctx, `
		SELECT balance FROM account_balances WHERE account_id = ?
	`, params.AccountID).Scan(&currentBalanceStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("account balance not found for %s", params.AccountID)
		}
		return nil, err
	}

	currentBalance, err := money.NewFromString(currentBalanceStr)
	if err != nil {
		return nil, err
	}

	var newBalance money.Money
	switch params.Direction {
	case model.DirectionIn:
		newBalance = currentBalance.Add(params.Amount)
	case model.DirectionOut:
		newBalance = currentBalance.Sub(params.Amount)
	default:
		return nil, fmt.Errorf("invalid direction: %s", params.Direction)
	}

	if newBalance.IsNegative() {
		return nil, fmt.Errorf("insufficient balance")
	}

	entryID := id.NewTransactionID()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO ledger_entries (id, account_id, transfer_id, biz_id, direction, type, amount, balance_after, counterparty, remark)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, entryID, params.AccountID, params.TransferID, params.BizID, params.Direction, params.Type,
		params.Amount.Amount.String(), newBalance.Amount.String(), params.Counterparty, params.Remark)
	if err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE account_balances 
		SET balance = ?, last_txn_id = ?, version = version + 1, updated_at = ?
		WHERE account_id = ?
	`, newBalance.Amount.String(), entryID, time.Now(), params.AccountID)
	if err != nil {
		return nil, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, fmt.Errorf("failed to update account balance")
	}

	result, err = tx.ExecContext(ctx, `
		UPDATE accounts 
		SET balance = ?, updated_at = ?
		WHERE id = ?
	`, newBalance.Amount.String(), time.Now(), params.AccountID)
	if err != nil {
		return nil, err
	}

	return &model.LedgerEntry{
		ID:           entryID,
		AccountID:    params.AccountID,
		TransferID:   params.TransferID,
		BizID:        params.BizID,
		Direction:    params.Direction,
		Type:         params.Type,
		Amount:       params.Amount,
		BalanceAfter: newBalance,
		Counterparty: params.Counterparty,
		Remark:       params.Remark,
	}, nil
}

func (s *Service) GetEntry(ctx context.Context, entryID string) (*model.LedgerEntry, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, account_id, transfer_id, biz_id, direction, type, amount, balance_after, counterparty, remark, created_at
		FROM ledger_entries WHERE id = ?
	`, entryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	return db.ScanLedgerEntry(rows)
}

func (s *Service) ListByAccount(ctx context.Context, accountID string, limit, offset int) ([]*model.LedgerEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, account_id, transfer_id, biz_id, direction, type, amount, balance_after, counterparty, remark, created_at
		FROM ledger_entries 
		WHERE account_id = ? 
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, accountID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*model.LedgerEntry
	for rows.Next() {
		e, err := db.ScanLedgerEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func (s *Service) GetBalance(ctx context.Context, accountID string) (money.Money, error) {
	var balanceStr string
	err := s.db.QueryRowContext(ctx, `
		SELECT balance FROM account_balances WHERE account_id = ?
	`, accountID).Scan(&balanceStr)
	if err != nil {
		return money.Zero(), err
	}
	return money.NewFromString(balanceStr)
}

func (s *Service) ReconcileAccount(ctx context.Context, accountID string, checkDate string) (*model.ReconciliationDiff, error) {
	var balanceStr string
	err := s.db.QueryRowContext(ctx, `
		SELECT balance FROM account_balances WHERE account_id = ?
	`, accountID).Scan(&balanceStr)
	if err != nil {
		return nil, err
	}
	balance, _ := money.NewFromString(balanceStr)

	var firstEntryID string
	err = s.db.QueryRowContext(ctx, `
		SELECT id FROM ledger_entries WHERE account_id = ? ORDER BY created_at ASC LIMIT 1
	`, accountID).Scan(&firstEntryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, account_id, transfer_id, biz_id, direction, type, amount, balance_after, counterparty, remark, created_at
		FROM ledger_entries 
		WHERE account_id = ? 
		ORDER BY created_at ASC
	`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var computedBalance = money.Zero()
	var lastBalanceAfter = money.Zero()
	for rows.Next() {
		e, err := db.ScanLedgerEntry(rows)
		if err != nil {
			return nil, err
		}
		if e.Direction == model.DirectionIn {
			computedBalance = computedBalance.Add(e.Amount)
		} else {
			computedBalance = computedBalance.Sub(e.Amount)
		}
		lastBalanceAfter = e.BalanceAfter
	}

	diff := balance.Sub(computedBalance)

	if diff.IsZero() && lastBalanceAfter.Equal(balance) {
		return nil, nil
	}

	diffID := id.NewReportID()
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO reconciliation_diffs (id, account_id, check_date, balance_amount, ledger_amount, diff_amount)
		VALUES (?, ?, ?, ?, ?, ?)
	`, diffID, accountID, checkDate, balance.Amount.String(), computedBalance.Amount.String(), diff.Amount.String())
	if err != nil {
		log.Printf("failed to record recon diff: %v", err)
	}

	return &model.ReconciliationDiff{
		ID:            diffID,
		AccountID:     accountID,
		CheckDate:     checkDate,
		BalanceAmount: balance,
		LedgerAmount:  computedBalance,
		DiffAmount:    diff,
	}, nil
}

func (s *Service) InitAccountBalance(ctx context.Context, tx *sql.Tx, accountID string, initialBalance money.Money) error {
	balanceID := id.NewAccountID() + "-bal"
	_, err := tx.ExecContext(ctx, `
		INSERT INTO account_balances (id, account_id, balance, frozen_amount, version)
		VALUES (?, ?, ?, 0, 0)
	`, balanceID, accountID, initialBalance.Amount.String())
	return err
}
