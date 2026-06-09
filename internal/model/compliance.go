package model

import "time"

type ReportType string

const (
	ReportTypeLargeAmount ReportType = "large_amount"
	ReportTypeSuspicious  ReportType = "suspicious"
)

type ReportStatus string

const (
	ReportStatusPending  ReportStatus = "pending"
	ReportStatusReported ReportStatus = "reported"
	ReportStatusFailed   ReportStatus = "failed"
)

type ComplianceReport struct {
	ID             string       `db:"id" json:"id"`
	ReportType     ReportType   `db:"report_type" json:"report_type"`
	AccountID      string       `db:"account_id" json:"account_id"`
	TransferID     string       `db:"transfer_id" json:"transfer_id"`
	ReportContent  string       `db:"report_content" json:"report_content"`
	Status         ReportStatus `db:"status" json:"status"`
	RetryCount     int          `db:"retry_count" json:"retry_count"`
	ReportBatchNo  string       `db:"report_batch_no" json:"report_batch_no"`
	ReportedAt     *time.Time   `db:"reported_at" json:"reported_at"`
	FailureReason  string       `db:"failure_reason" json:"failure_reason"`
	CreatedAt      time.Time    `db:"created_at" json:"created_at"`
}
