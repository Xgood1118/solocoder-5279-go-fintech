package model

import "time"

type AuditAction string

const (
	AuditActionCreateAccount  AuditAction = "create_account"
	AuditActionCloseAccount   AuditAction = "close_account"
	AuditActionFreezeAccount  AuditAction = "freeze_account"
	AuditActionUnfreezeAccount AuditAction = "unfreeze_account"
	AuditActionChangePassword AuditAction = "change_password"
	AuditActionTransfer       AuditAction = "transfer"
	AuditActionCrossBankTransfer AuditAction = "cross_bank_transfer"
	AuditActionInterestSettle AuditAction = "interest_settle"
)

type AuditLog struct {
	ID         string      `db:"id" json:"id"`
	Action     AuditAction `db:"action" json:"action"`
	AccountID  string      `db:"account_id" json:"account_id"`
	OperatorID string      `db:"operator_id" json:"operator_id"`
	IP         string      `db:"ip" json:"ip"`
	UserAgent  string      `db:"user_agent" json:"user_agent"`
	RequestID  string      `db:"request_id" json:"request_id"`
	Before     string      `db:"before" json:"before"`
	After      string      `db:"after" json:"after"`
	Detail     string      `db:"detail" json:"detail"`
	CreatedAt  time.Time   `db:"created_at" json:"created_at"`
}
