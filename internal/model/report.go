package model

import (
	"time"

	"github.com/fintech/core/pkg/money"
)

type DailyAccountReport struct {
	ID            string      `db:"id" json:"id"`
	AccountID     string      `db:"account_id" json:"account_id"`
	ReportDate    string      `db:"report_date" json:"report_date"`
	BeginBalance  money.Money `db:"begin_balance" json:"begin_balance"`
	EndBalance    money.Money `db:"end_balance" json:"end_balance"`
	TotalIncome   money.Money `db:"total_income" json:"total_income"`
	TotalExpense  money.Money `db:"total_expense" json:"total_expense"`
	TxnCount      int         `db:"txn_count" json:"txn_count"`
	CreatedAt     time.Time   `db:"created_at" json:"created_at"`
}

type MonthlyAccountReport struct {
	ID            string      `db:"id" json:"id"`
	AccountID     string      `db:"account_id" json:"account_id"`
	ReportMonth   string      `db:"report_month" json:"report_month"`
	BeginBalance  money.Money `db:"begin_balance" json:"begin_balance"`
	EndBalance    money.Money `db:"end_balance" json:"end_balance"`
	TotalIncome   money.Money `db:"total_income" json:"total_income"`
	TotalExpense  money.Money `db:"total_expense" json:"total_expense"`
	Interest      money.Money `db:"interest" json:"interest"`
	TxnCount      int         `db:"txn_count" json:"txn_count"`
	CreatedAt     time.Time   `db:"created_at" json:"created_at"`
}

type ReconciliationDiff struct {
	ID              string      `db:"id" json:"id"`
	AccountID       string      `db:"account_id" json:"account_id"`
	CheckDate       string      `db:"check_date" json:"check_date"`
	BalanceAmount   money.Money `db:"balance_amount" json:"balance_amount"`
	LedgerAmount    money.Money `db:"ledger_amount" json:"ledger_amount"`
	DiffAmount      money.Money `db:"diff_amount" json:"diff_amount"`
	IsFixed         bool        `db:"is_fixed" json:"is_fixed"`
	FixedAt         *time.Time  `db:"fixed_at" json:"fixed_at"`
	CreatedAt       time.Time   `db:"created_at" json:"created_at"`
}
