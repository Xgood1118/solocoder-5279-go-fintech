package model

import "time"

type BlacklistType string

const (
	BlacklistTypeAccount BlacklistType = "account"
	BlacklistTypeIDCard  BlacklistType = "id_card"
	BlacklistTypeIP      BlacklistType = "ip"
)

type BlacklistEntry struct {
	ID        string        `db:"id" json:"id"`
	Type      BlacklistType `db:"type" json:"type"`
	Value     string        `db:"value" json:"value"`
	Reason    string        `db:"reason" json:"reason"`
	ExpiresAt *time.Time    `db:"expires_at" json:"expires_at"`
	CreatedAt time.Time     `db:"created_at" json:"created_at"`
}

type RiskLimitType string

const (
	RiskLimitSingle  RiskLimitType = "single"
	RiskLimitDaily   RiskLimitType = "daily"
	RiskLimitMonthly RiskLimitType = "monthly"
)

type SuspiciousType string

const (
	SuspiciousTypeFrequent      SuspiciousType = "frequent"
	SuspiciousTypeLargeAmount   SuspiciousType = "large_amount"
	SuspiciousTypeAbnormalPattern SuspiciousType = "abnormal_pattern"
)

type SuspiciousTransaction struct {
	ID           string         `db:"id" json:"id"`
	AccountID    string         `db:"account_id" json:"account_id"`
	TransferID   string         `db:"transfer_id" json:"transfer_id"`
	Type         SuspiciousType `db:"type" json:"type"`
	Description  string         `db:"description" json:"description"`
	RiskScore    int            `db:"risk_score" json:"risk_score"`
	Reported     bool           `db:"reported" json:"reported"`
	ReportedAt   *time.Time     `db:"reported_at" json:"reported_at"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
}
