package account

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/fintech/core/internal/db"
	"github.com/fintech/core/internal/ledger"
	"github.com/fintech/core/internal/model"
	"github.com/fintech/core/pkg/id"
	"github.com/fintech/core/pkg/money"
	fmterrors "github.com/fintech/core/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	db      *db.DB
	ledger  *ledger.Service
	hashCost int
}

type CreateAccountParams struct {
	AccountType model.AccountType
	Name        string
	IDCardNo    string
	Password    string
	InitialBalance money.Money
}

func NewService(database *db.DB, ledgerSvc *ledger.Service, hashCost int) *Service {
	if hashCost <= 0 {
		hashCost = bcrypt.DefaultCost
	}
	return &Service{
		db:       database,
		ledger:   ledgerSvc,
		hashCost: hashCost,
	}
}

func (s *Service) Create(ctx context.Context, params CreateAccountParams) (*model.Account, error) {
	if params.Name == "" {
		return nil, fmterrors.NewBadRequest("INVALID_NAME", "姓名不能为空")
	}
	if params.Password == "" || len(params.Password) < 6 {
		return nil, fmterrors.NewBadRequest("INVALID_PASSWORD", "密码长度不能少于6位")
	}

	accountID := id.NewAccountID()
	accountNo := generateAccountNo()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(params.Password), s.hashCost)
	if err != nil {
		return nil, fmterrors.Wrap(err, "密码加密失败")
	}

	err = s.db.WithTx(ctx, func(tx *sql.Tx) error {
		now := time.Now()
		_, err := tx.ExecContext(ctx, `
			INSERT INTO accounts (id, account_type, account_no, name, id_card_no, password_hash, status, balance, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, 'active', '0', ?, ?)
		`, accountID, params.AccountType, accountNo, params.Name, params.IDCardNo,
			string(passwordHash), now, now)
		if err != nil {
			return err
		}

		if err := s.ledger.InitAccountBalance(ctx, tx, accountID, money.Zero()); err != nil {
			return err
		}

		if !params.InitialBalance.IsZero() {
			_, err := s.ledger.CreateEntry(ctx, tx, ledger.CreateEntryParams{
				AccountID: accountID,
				BizID:     accountID + "-init",
				Direction: model.DirectionIn,
				Type:      model.TxnTypeDeposit,
				Amount:    params.InitialBalance,
				Remark:    "开户初始存入",
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.Get(ctx, accountID)
}

func (s *Service) Get(ctx context.Context, accountID string) (*model.Account, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, account_type, account_no, name, id_card_no, password_hash, status, balance, frozen_amount, created_at, updated_at, closed_at
		FROM accounts WHERE id = ?
	`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmterrors.ErrAccountNotFound
	}
	return db.ScanAccount(rows)
}

func (s *Service) GetByAccountNo(ctx context.Context, accountNo string) (*model.Account, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, account_type, account_no, name, id_card_no, password_hash, status, balance, frozen_amount, created_at, updated_at, closed_at
		FROM accounts WHERE account_no = ?
	`, accountNo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmterrors.ErrAccountNotFound
	}
	return db.ScanAccount(rows)
}

func (s *Service) GetBalance(ctx context.Context, accountID string) (money.Money, error) {
	return s.ledger.GetBalance(ctx, accountID)
}

func (s *Service) Freeze(ctx context.Context, accountID string) error {
	account, err := s.Get(ctx, accountID)
	if err != nil {
		return err
	}
	if account.Status == model.AccountStatusClosed {
		return fmterrors.ErrAccountClosed
	}
	if account.Status == model.AccountStatusFrozen {
		return nil
	}

	result, err := s.db.ExecContext(ctx, `
		UPDATE accounts SET status = 'frozen', updated_at = ? WHERE id = ?
	`, time.Now(), accountID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmterrors.ErrAccountNotFound
	}
	return nil
}

func (s *Service) Unfreeze(ctx context.Context, accountID string) error {
	account, err := s.Get(ctx, accountID)
	if err != nil {
		return err
	}
	if account.Status == model.AccountStatusClosed {
		return fmterrors.ErrAccountClosed
	}
	if account.Status == model.AccountStatusActive {
		return nil
	}

	result, err := s.db.ExecContext(ctx, `
		UPDATE accounts SET status = 'active', updated_at = ? WHERE id = ?
	`, time.Now(), accountID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmterrors.ErrAccountNotFound
	}
	return nil
}

func (s *Service) Close(ctx context.Context, accountID string) error {
	account, err := s.Get(ctx, accountID)
	if err != nil {
		return err
	}
	if account.Status == model.AccountStatusClosed {
		return nil
	}
	if !account.Balance.IsZero() {
		return fmterrors.NewBadRequest("BALANCE_NOT_ZERO", "账户余额不为零，不能销户")
	}

	now := time.Now()
	result, err := s.db.ExecContext(ctx, `
		UPDATE accounts SET status = 'closed', closed_at = ?, updated_at = ? WHERE id = ?
	`, now, now, accountID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmterrors.ErrAccountNotFound
	}
	return nil
}

func (s *Service) ChangePassword(ctx context.Context, accountID string, oldPassword, newPassword string) error {
	account, err := s.Get(ctx, accountID)
	if err != nil {
		return err
	}
	if account.Status == model.AccountStatusClosed {
		return fmterrors.ErrAccountClosed
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(oldPassword)); err != nil {
		return fmterrors.ErrInvalidPassword
	}
	if len(newPassword) < 6 {
		return fmterrors.NewBadRequest("INVALID_NEW_PASSWORD", "新密码长度不能少于6位")
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.hashCost)
	if err != nil {
		return fmterrors.Wrap(err, "密码加密失败")
	}

	result, err := s.db.ExecContext(ctx, `
		UPDATE accounts SET password_hash = ?, updated_at = ? WHERE id = ?
	`, string(newHash), time.Now(), accountID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmterrors.ErrAccountNotFound
	}
	return nil
}

func (s *Service) VerifyPassword(ctx context.Context, accountID string, password string) (bool, error) {
	account, err := s.Get(ctx, accountID)
	if err != nil {
		return false, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *Service) CheckAvailable(ctx context.Context, accountID string, amount money.Money) error {
	account, err := s.Get(ctx, accountID)
	if err != nil {
		return err
	}

	switch account.Status {
	case model.AccountStatusFrozen:
		return fmterrors.ErrAccountFrozen
	case model.AccountStatusClosed:
		return fmterrors.ErrAccountClosed
	}

	if account.Balance.LessThan(amount) {
		return fmterrors.ErrInsufficientBalance
	}

	return nil
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]*model.Account, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, account_type, account_no, name, id_card_no, password_hash, status, balance, frozen_amount, created_at, updated_at, closed_at
		FROM accounts ORDER BY created_at DESC LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*model.Account
	for rows.Next() {
		a, err := db.ScanAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

func generateAccountNo() string {
	buf := make([]byte, 6)
	_, _ = rand.Read(buf)
	return fmt.Sprintf("6222%02d%02d%02d%04d",
		time.Now().Year()%100, int(time.Now().Month()), time.Now().Day(),
		uint16(buf[0])<<8|uint16(buf[1]))
}
