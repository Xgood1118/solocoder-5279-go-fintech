package model

import (
	"time"

	"github.com/fintech/core/pkg/money"
)

type TransactionDirection string

const (
	DirectionIn  TransactionDirection = "in"
	DirectionOut TransactionDirection = "out"
)

type TransactionType string

const (
	TxnTypeDeposit      TransactionType = "deposit"
	TxnTypeWithdraw     TransactionType = "withdraw"
	TxnTypeTransferIn   TransactionType = "transfer_in"
	TxnTypeTransferOut  TransactionType = "transfer_out"
	TxnTypeInterest     TransactionType = "interest"
	TxnTypeFee          TransactionType = "fee"
	TxnTypeFreeze       TransactionType = "freeze"
	TxnTypeUnfreeze     TransactionType = "unfreeze"
	TxnTypeCrossBankIn  TransactionType = "cross_bank_in"
	TxnTypeCrossBankOut TransactionType = "cross_bank_out"
)

type LedgerEntry struct {
	ID             string               `db:"id" json:"id"`
	AccountID      string               `db:"account_id" json:"account_id"`
	TransferID     string               `db:"transfer_id" json:"transfer_id"`
	BizID          string               `db:"biz_id" json:"biz_id"`
	Direction      TransactionDirection `db:"direction" json:"direction"`
	Type           TransactionType      `db:"type" json:"type"`
	Amount         money.Money          `db:"amount" json:"amount"`
	BalanceAfter   money.Money          `db:"balance_after" json:"balance_after"`
	Counterparty   string               `db:"counterparty" json:"counterparty"`
	Remark         string               `db:"remark" json:"remark"`
	CreatedAt      time.Time            `db:"created_at" json:"created_at"`
}

type TransferStatus string

const (
	TransferStatusPending   TransferStatus = "pending"
	TransferStatusSuccess   TransferStatus = "success"
	TransferStatusFailed    TransferStatus = "failed"
	TransferStatusReversed  TransferStatus = "reversed"
	TransferStatusSettling  TransferStatus = "settling"
)

type Transfer struct {
	ID             string         `db:"id" json:"id"`
	BizID          string         `db:"biz_id" json:"biz_id"`
	FromAccountID  string         `db:"from_account_id" json:"from_account_id"`
	ToAccountID    string         `db:"to_account_id" json:"to_account_id"`
	ToBankCode     string         `db:"to_bank_code" json:"to_bank_code"`
	ToAccountNo    string         `db:"to_account_no" json:"to_account_no"`
	ToAccountName  string         `db:"to_account_name" json:"to_account_name"`
	Amount         money.Money    `db:"amount" json:"amount"`
	Fee            money.Money    `db:"fee" json:"fee"`
	Remark         string         `db:"remark" json:"remark"`
	Status         TransferStatus `db:"status" json:"status"`
	IsCrossBank    bool           `db:"is_cross_bank" json:"is_cross_bank"`
	SettlementID   string         `db:"settlement_id" json:"settlement_id"`
	FailureReason  string         `db:"failure_reason" json:"failure_reason"`
	CreatedAt      time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at" json:"updated_at"`
	CompletedAt    *time.Time     `db:"completed_at" json:"completed_at"`
}
