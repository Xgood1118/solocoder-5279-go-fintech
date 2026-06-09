package interest

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/fintech/core/internal/config"
	"github.com/fintech/core/internal/db"
	"github.com/fintech/core/internal/ledger"
	"github.com/fintech/core/internal/model"
	"github.com/fintech/core/pkg/id"
	"github.com/fintech/core/pkg/money"
	"github.com/shopspring/decimal"
)

type Service struct {
	db     *db.DB
	ledger *ledger.Service
	config *config.InterestConfig
}

func NewService(database *db.DB, ledgerSvc *ledger.Service, cfg *config.InterestConfig) *Service {
	return &Service{
		db:     database,
		ledger: ledgerSvc,
		config: cfg,
	}
}

func (s *Service) AccrueDailyInterest(ctx context.Context, accountID string, accrualDate string) error {
	balance, err := s.ledger.GetBalance(ctx, accountID)
	if err != nil {
		return err
	}

	dailyRate := s.config.DemandRate.Div(decimal.NewFromInt(360))
	accrued := balance.MulDecimal(dailyRate)

	id := id.NewReportID()
	_, err = s.db.ExecContext(ctx, `
		INSERT OR IGNORE INTO interest_accruals (id, account_id, accrual_date, principal, accrued)
		VALUES (?, ?, ?, ?, ?)
	`, id, accountID, accrualDate, balance.Amount.String(), accrued.Amount.String())
	return err
}

func (s *Service) SettleMonthlyInterest(ctx context.Context, accountID string, year int, month time.Month) error {
	periodStart := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	periodEnd := time.Date(year, month+1, 1, 0, 0, 0, -1, time.Local)

	var totalPrincipalStr, totalAccruedStr string
	err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(principal), '0'), COALESCE(SUM(accrued), '0')
		FROM interest_accruals
		WHERE account_id = ? AND accrual_date >= ? AND accrual_date <= ?
	`, accountID, periodStart.Format("2006-01-02"), periodEnd.Format("2006-01-02")).Scan(&totalPrincipalStr, &totalAccruedStr)
	if err != nil {
		return err
	}

	totalAccrued, _ := money.NewFromString(totalAccruedStr)
	totalPrincipal, _ := money.NewFromString(totalPrincipalStr)

	if totalAccrued.IsZero() {
		return nil
	}

	err = s.db.WithTx(ctx, func(tx *sql.Tx) error {
		_, err := s.ledger.CreateEntry(ctx, tx, ledger.CreateEntryParams{
			AccountID: accountID,
			Direction: model.DirectionIn,
			Type:      model.TxnTypeInterest,
			Amount:    totalAccrued,
			Remark:    month.String() + " 活期利息结息",
		})
		if err != nil {
			return err
		}

		recordID := id.NewReportID()
		_, err = tx.ExecContext(ctx, `
			INSERT INTO interest_records (id, account_id, ledger_id, interest_type, principal, rate, days, interest, period_start, period_end)
			VALUES (?, ?, ?, 'demand', ?, ?, ?, ?, ?, ?)
		`, recordID, accountID, "", "", totalPrincipal.Amount.String(), s.config.DemandRate.String(),
			daysInMonth(year, month), totalAccrued.Amount.String(), periodStart, periodEnd)
		return err
	})

	return err
}

func (s *Service) CreateFixedDeposit(ctx context.Context, accountID string, principal money.Money, interestType model.InterestType, autoRenew bool) (*model.FixedDeposit, error) {
	var rate decimal.Decimal
	var duration time.Duration

	switch interestType {
	case model.InterestTypeFixed3M:
		rate = s.config.Fixed3MRate
		duration = 3 * 30 * 24 * time.Hour
	case model.InterestTypeFixed6M:
		rate = s.config.Fixed6MRate
		duration = 6 * 30 * 24 * time.Hour
	case model.InterestTypeFixed1Y:
		rate = s.config.Fixed1YRate
		duration = 365 * 24 * time.Hour
	default:
		return nil, nil
	}

	interest := principal.MulDecimal(rate)

	id := id.NewReportID()
	startDate := time.Now()
	maturityDate := startDate.Add(duration)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO fixed_deposits (id, account_id, principal, interest_type, annual_rate, interest_amount, start_date, maturity_date, auto_renew)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, accountID, principal.Amount.String(), interestType, rate.String(), interest.Amount.String(),
		startDate, maturityDate, autoRenew)
	if err != nil {
		return nil, err
	}

	return &model.FixedDeposit{
		ID:             id,
		AccountID:      accountID,
		Principal:      principal,
		InterestType:   interestType,
		AnnualRate:     rate.String(),
		InterestAmount: interest,
		StartDate:      startDate,
		MaturityDate:   maturityDate,
		AutoRenew:      autoRenew,
	}, nil
}

func (s *Service) ProcessMaturedDeposits(ctx context.Context) (int, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, account_id, principal, interest_type, annual_rate, interest_amount, start_date, maturity_date, auto_renew, is_redeemed, redeemed_at, created_at
		FROM fixed_deposits 
		WHERE maturity_date <= ? AND is_redeemed = 0
	`, time.Now())
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var deposits []*model.FixedDeposit
	for rows.Next() {
		d, err := db.ScanFixedDeposit(rows)
		if err != nil {
			return 0, err
		}
		deposits = append(deposits, d)
	}

	count := 0
	for _, d := range deposits {
		if err := s.redeemFixedDeposit(ctx, d); err != nil {
			log.Printf("Redeem fixed deposit %s failed: %v", d.ID, err)
			continue
		}
		count++
	}

	return count, nil
}

func (s *Service) redeemFixedDeposit(ctx context.Context, deposit *model.FixedDeposit) error {
	return s.db.WithTx(ctx, func(tx *sql.Tx) error {
		_, err := s.ledger.CreateEntry(ctx, tx, ledger.CreateEntryParams{
			AccountID: deposit.AccountID,
			Direction: model.DirectionIn,
			Type:      model.TxnTypeInterest,
			Amount:    deposit.InterestAmount,
			Remark:    "定期存款到期利息",
		})
		if err != nil {
			return err
		}

		now := time.Now()
		_, err = tx.ExecContext(ctx, `
			UPDATE fixed_deposits SET is_redeemed = 1, redeemed_at = ? WHERE id = ?
		`, now, deposit.ID)
		if err != nil {
			return err
		}

		if deposit.AutoRenew {
			var newRate decimal.Decimal
			var newDuration time.Duration

			switch deposit.InterestType {
			case model.InterestTypeFixed3M:
				newRate = s.config.Fixed3MRate
				newDuration = 3 * 30 * 24 * time.Hour
			case model.InterestTypeFixed6M:
				newRate = s.config.Fixed6MRate
				newDuration = 6 * 30 * 24 * time.Hour
			case model.InterestTypeFixed1Y:
				newRate = s.config.Fixed1YRate
				newDuration = 365 * 24 * time.Hour
			}

			newInterest := deposit.Principal.MulDecimal(newRate)
			newID := id.NewReportID()
			newStart := now
			newMaturity := newStart.Add(newDuration)

			_, err = tx.ExecContext(ctx, `
				INSERT INTO fixed_deposits (id, account_id, principal, interest_type, annual_rate, interest_amount, start_date, maturity_date, auto_renew)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, newID, deposit.AccountID, deposit.Principal.Amount.String(), deposit.InterestType,
				newRate.String(), newInterest.Amount.String(), newStart, newMaturity, deposit.AutoRenew)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *Service) ListFixedDeposits(ctx context.Context, accountID string, limit, offset int) ([]*model.FixedDeposit, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, account_id, principal, interest_type, annual_rate, interest_amount, start_date, maturity_date, auto_renew, is_redeemed, redeemed_at, created_at
		FROM fixed_deposits WHERE account_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?
	`, accountID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deposits []*model.FixedDeposit
	for rows.Next() {
		d, err := db.ScanFixedDeposit(rows)
		if err != nil {
			return nil, err
		}
		deposits = append(deposits, d)
	}
	return deposits, nil
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
}
