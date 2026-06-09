package model

import (
	"time"

	"github.com/fintech/core/pkg/money"
)

type InterestType string

const (
	InterestTypeDemand  InterestType = "demand"
	InterestTypeFixed3M InterestType = "fixed_3m"
	InterestTypeFixed6M InterestType = "fixed_6m"
	InterestTypeFixed1Y InterestType = "fixed_1y"
)

type FixedDeposit struct {
	ID             string       `db:"id" json:"id"`
	AccountID      string       `db:"account_id" json:"account_id"`
	Principal      money.Money  `db:"principal" json:"principal"`
	InterestType   InterestType `db:"interest_type" json:"interest_type"`
	AnnualRate     string       `db:"annual_rate" json:"annual_rate"`
	InterestAmount money.Money  `db:"interest_amount" json:"interest_amount"`
	StartDate      time.Time    `db:"start_date" json:"start_date"`
	MaturityDate   time.Time    `db:"maturity_date" json:"maturity_date"`
	AutoRenew      bool         `db:"auto_renew" json:"auto_renew"`
	IsRedeemed     bool         `db:"is_redeemed" json:"is_redeemed"`
	RedeemedAt     *time.Time   `db:"redeemed_at" json:"redeemed_at"`
	CreatedAt      time.Time    `db:"created_at" json:"created_at"`
}

type InterestRecord struct {
	ID           string       `db:"id" json:"id"`
	AccountID    string       `db:"account_id" json:"account_id"`
	LedgerID     string       `db:"ledger_id" json:"ledger_id"`
	InterestType InterestType `db:"interest_type" json:"interest_type"`
	Principal    money.Money  `db:"principal" json:"principal"`
	Rate         string       `db:"rate" json:"rate"`
	Days         int          `db:"days" json:"days"`
	Interest     money.Money  `db:"interest" json:"interest"`
	PeriodStart  time.Time    `db:"period_start" json:"period_start"`
	PeriodEnd    time.Time    `db:"period_end" json:"period_end"`
	CreatedAt    time.Time    `db:"created_at" json:"created_at"`
}

type InterestAccrual struct {
	ID          string      `db:"id" json:"id"`
	AccountID   string      `db:"account_id" json:"account_id"`
	AccrualDate string      `db:"accrual_date" json:"accrual_date"`
	Principal   money.Money `db:"principal" json:"principal"`
	Accrued     money.Money `db:"accrued" json:"accrued"`
	CreatedAt   time.Time   `db:"created_at" json:"created_at"`
}
