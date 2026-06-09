package transfer

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/fintech/core/internal/account"
	"github.com/fintech/core/internal/db"
	"github.com/fintech/core/internal/ledger"
	"github.com/fintech/core/internal/model"
	"github.com/fintech/core/internal/risk"
	"github.com/fintech/core/pkg/id"
	"github.com/fintech/core/pkg/money"
	fmterrors "github.com/fintech/core/pkg/errors"
)

type Service struct {
	db      *db.DB
	account *account.Service
	ledger  *ledger.Service
	risk    *risk.Service
}

type IntraBankTransferParams struct {
	BizID         string
	FromAccountID string
	ToAccountID   string
	Amount        money.Money
	Remark        string
	Password      string
	IP            string
}

type CrossBankTransferParams struct {
	BizID         string
	FromAccountID string
	ToBankCode    string
	ToAccountNo   string
	ToAccountName string
	Amount        money.Money
	Remark        string
	Password      string
	IP            string
}

func NewService(database *db.DB, accountSvc *account.Service, ledgerSvc *ledger.Service, riskSvc *risk.Service) *Service {
	return &Service{
		db:      database,
		account: accountSvc,
		ledger:  ledgerSvc,
		risk:    riskSvc,
	}
}

func (s *Service) IntraBankTransfer(ctx context.Context, params IntraBankTransferParams) (*model.Transfer, error) {
	if params.FromAccountID == params.ToAccountID {
		return nil, fmterrors.ErrSameAccount
	}
	if params.Amount.IsZero() || params.Amount.IsNegative() {
		return nil, fmterrors.ErrInvalidAmount
	}

	existing, err := s.GetByBizID(ctx, params.BizID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if existing != nil {
		if existing.Status == model.TransferStatusSuccess {
			return existing, nil
		}
		return nil, fmterrors.ErrIdempotentConflict
	}

	riskParams := risk.CheckTransferParams{
		FromAccountID: params.FromAccountID,
		ToAccountID:   params.ToAccountID,
		Amount:        params.Amount,
		IP:            params.IP,
	}
	if err := s.risk.CheckTransfer(ctx, riskParams); err != nil {
		return nil, err
	}

	ok, err := s.account.VerifyPassword(ctx, params.FromAccountID, params.Password)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmterrors.ErrInvalidPassword
	}

	fromAccount, err := s.account.Get(ctx, params.FromAccountID)
	if err != nil {
		return nil, err
	}
	if fromAccount.Status == model.AccountStatusFrozen {
		return nil, fmterrors.ErrAccountFrozen
	}
	if fromAccount.Status == model.AccountStatusClosed {
		return nil, fmterrors.ErrAccountClosed
	}

	toAccount, err := s.account.Get(ctx, params.ToAccountID)
	if err != nil {
		return nil, err
	}
	if toAccount.Status == model.AccountStatusClosed {
		return nil, fmterrors.NewForbidden("TARGET_ACCOUNT_CLOSED", "收款账户已销户")
	}

	transferID := id.NewTransferID()

	err = s.db.WithTx(ctx, func(tx *sql.Tx) error {
		now := time.Now()
		_, err := tx.ExecContext(ctx, `
			INSERT INTO transfers (id, biz_id, from_account_id, to_account_id, amount, remark, status, is_cross_bank, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, 'pending', 0, ?, ?)
		`, transferID, params.BizID, params.FromAccountID, params.ToAccountID,
			params.Amount.Amount.String(), params.Remark, now, now)
		if err != nil {
			return err
		}

		_, err = s.ledger.CreateEntry(ctx, tx, ledger.CreateEntryParams{
			AccountID:    params.FromAccountID,
			TransferID:   transferID,
			BizID:        params.BizID,
			Direction:    model.DirectionOut,
			Type:         model.TxnTypeTransferOut,
			Amount:       params.Amount,
			Counterparty: toAccount.Name,
			Remark:       params.Remark,
		})
		if err != nil {
			return err
		}

		_, err = s.ledger.CreateEntry(ctx, tx, ledger.CreateEntryParams{
			AccountID:    params.ToAccountID,
			TransferID:   transferID,
			BizID:        params.BizID,
			Direction:    model.DirectionIn,
			Type:         model.TxnTypeTransferIn,
			Amount:       params.Amount,
			Counterparty: fromAccount.Name,
			Remark:       params.Remark,
		})
		if err != nil {
			return err
		}

		now = time.Now()
		result, err := tx.ExecContext(ctx, `
			UPDATE transfers SET status = 'success', completed_at = ?, updated_at = ?
			WHERE id = ?
		`, now, now, transferID)
		if err != nil {
			return err
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			return errors.New("failed to update transfer status")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	go func() {
		_, _ = s.risk.DetectSuspicious(context.Background(), params.FromAccountID, params.Amount)
	}()

	return s.Get(ctx, transferID)
}

func (s *Service) CrossBankTransfer(ctx context.Context, params CrossBankTransferParams) (*model.Transfer, error) {
	if params.Amount.IsZero() || params.Amount.IsNegative() {
		return nil, fmterrors.ErrInvalidAmount
	}

	existing, err := s.GetByBizID(ctx, params.BizID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	riskParams := risk.CheckTransferParams{
		FromAccountID: params.FromAccountID,
		ToAccountID:   params.ToAccountNo,
		Amount:        params.Amount,
		IP:            params.IP,
	}
	if err := s.risk.CheckTransfer(ctx, riskParams); err != nil {
		return nil, err
	}

	ok, err := s.account.VerifyPassword(ctx, params.FromAccountID, params.Password)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmterrors.ErrInvalidPassword
	}

	if err := s.account.CheckAvailable(ctx, params.FromAccountID, params.Amount); err != nil {
		return nil, err
	}

	transferID := id.NewTransferID()
	settlementID := id.NewSettlementID()

	err = s.db.WithTx(ctx, func(tx *sql.Tx) error {
		now := time.Now()
		_, err := tx.ExecContext(ctx, `
			INSERT INTO transfers (id, biz_id, from_account_id, to_bank_code, to_account_no, to_account_name,
				amount, remark, status, is_cross_bank, settlement_id, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'settling', 1, ?, ?, ?)
		`, transferID, params.BizID, params.FromAccountID, params.ToBankCode, params.ToAccountNo,
			params.ToAccountName, params.Amount.Amount.String(), params.Remark, settlementID, now, now)
		if err != nil {
			return err
		}

		_, err = s.ledger.CreateEntry(ctx, tx, ledger.CreateEntryParams{
			AccountID:    params.FromAccountID,
			TransferID:   transferID,
			BizID:        params.BizID,
			Direction:    model.DirectionOut,
			Type:         model.TxnTypeCrossBankOut,
			Amount:       params.Amount,
			Counterparty: params.ToAccountName,
			Remark:       params.Remark + " (跨行清算中)",
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	go s.processCrossBankSettlement(transferID, settlementID)

	return s.Get(ctx, transferID)
}

func (s *Service) processCrossBankSettlement(transferID, settlementID string) {
	ctx := context.Background()
	log.Printf("Processing cross-bank settlement for transfer %s", transferID)

	transfer, err := s.Get(ctx, transferID)
	if err != nil {
		log.Printf("Get transfer %s failed: %v", transferID, err)
		return
	}

	success := mockCentralBankSettlement(settlementID, transfer.Amount)

	if success {
		now := time.Now()
		_, err := s.db.ExecContext(ctx, `
			UPDATE transfers SET status = 'success', completed_at = ?, updated_at = ? WHERE id = ?
		`, now, now, transferID)
		if err != nil {
			log.Printf("Update transfer status failed: %v", err)
		}
		log.Printf("Cross-bank transfer %s completed successfully", transferID)
	} else {
		s.rollbackCrossBankTransfer(ctx, transferID, "央行清算失败")
	}
}

func (s *Service) rollbackCrossBankTransfer(ctx context.Context, transferID string, reason string) {
	transfer, err := s.Get(ctx, transferID)
	if err != nil {
		log.Printf("Get transfer for rollback failed: %v", err)
		return
	}

	err = s.db.WithTx(ctx, func(tx *sql.Tx) error {
		_, err := s.ledger.CreateEntry(ctx, tx, ledger.CreateEntryParams{
			AccountID:    transfer.FromAccountID,
			TransferID:   transferID,
			BizID:        transfer.BizID + "-rev",
			Direction:    model.DirectionIn,
			Type:         model.TxnTypeCrossBankOut,
			Amount:       transfer.Amount,
			Counterparty: "系统",
			Remark:       "跨行转账失败退回: " + reason,
		})
		if err != nil {
			return err
		}

		now := time.Now()
		_, err = tx.ExecContext(ctx, `
			UPDATE transfers SET status = 'failed', failure_reason = ?, updated_at = ? WHERE id = ?
		`, reason, now, transferID)
		return err
	})
	if err != nil {
		log.Printf("Rollback cross-bank transfer %s failed: %v", transferID, err)
	}
}

func mockCentralBankSettlement(settlementID string, amount money.Money) bool {
	log.Printf("Mock central bank settlement: %s, amount: %s", settlementID, amount.String())
	time.Sleep(500 * time.Millisecond)
	return true
}

func (s *Service) Get(ctx context.Context, transferID string) (*model.Transfer, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, biz_id, from_account_id, to_account_id, to_bank_code, to_account_no, to_account_name,
			amount, fee, remark, status, is_cross_bank, settlement_id, failure_reason, created_at, updated_at, completed_at
		FROM transfers WHERE id = ?
	`, transferID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	return db.ScanTransfer(rows)
}

func (s *Service) GetByBizID(ctx context.Context, bizID string) (*model.Transfer, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, biz_id, from_account_id, to_account_id, to_bank_code, to_account_no, to_account_name,
			amount, fee, remark, status, is_cross_bank, settlement_id, failure_reason, created_at, updated_at, completed_at
		FROM transfers WHERE biz_id = ?
	`, bizID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	return db.ScanTransfer(rows)
}

func (s *Service) ListByAccount(ctx context.Context, accountID string, limit, offset int) ([]*model.Transfer, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, biz_id, from_account_id, to_account_id, to_bank_code, to_account_no, to_account_name,
			amount, fee, remark, status, is_cross_bank, settlement_id, failure_reason, created_at, updated_at, completed_at
		FROM transfers 
		WHERE from_account_id = ? OR to_account_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, accountID, accountID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transfers []*model.Transfer
	for rows.Next() {
		t, err := db.ScanTransfer(rows)
		if err != nil {
			return nil, err
		}
		transfers = append(transfers, t)
	}
	return transfers, nil
}
