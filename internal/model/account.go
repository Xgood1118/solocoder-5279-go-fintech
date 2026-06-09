package model

import (
	"time"

	"github.com/fintech/core/pkg/money"
)

type AccountType string

const (
	AccountTypePersonal AccountType = "personal"
	AccountTypeEnterprise AccountType = "enterprise"
)

type AccountStatus string

const (
	AccountStatusActive  AccountStatus = "active"
	AccountStatusFrozen  AccountStatus = "frozen"
	AccountStatusClosed  AccountStatus = "closed"
)

type Account struct {
	ID              string        `db:"id" json:"id"`
	AccountType     AccountType   `db:"account_type" json:"account_type"`
	AccountNo       string        `db:"account_no" json:"account_no"`
	Name            string        `db:"name" json:"name"`
	IDCardNo        string        `db:"id_card_no" json:"id_card_no,omitempty"`
	PasswordHash    string        `db:"password_hash" json:"-"`
	Status          AccountStatus `db:"status" json:"status"`
	Balance         money.Money   `db:"balance" json:"balance"`
	FrozenAmount    money.Money   `db:"frozen_amount" json:"frozen_amount"`
	CreatedAt       time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time     `db:"updated_at" json:"updated_at"`
	ClosedAt        *time.Time    `db:"closed_at" json:"closed_at,omitempty"`
}

type AccountBalance struct {
	ID            string      `db:"id" json:"id"`
	AccountID     string      `db:"account_id" json:"account_id"`
	Balance       money.Money `db:"balance" json:"balance"`
	FrozenAmount  money.Money `db:"frozen_amount" json:"frozen_amount"`
	Version       int64       `db:"version" json:"version"`
	LastTxnID     string      `db:"last_txn_id" json:"last_txn_id"`
	UpdatedAt     time.Time   `db:"updated_at" json:"updated_at"`
}
