package id

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

func NewAccountID() string {
	return "ACC" + newID()
}

func NewTransactionID() string {
	return "TXN" + newID()
}

func NewTransferID() string {
	return "TRF" + newID()
}

func NewAuditID() string {
	return "AUD" + newID()
}

func NewReportID() string {
	return "RPT" + newID()
}

func NewSettlementID() string {
	return "STL" + newID()
}

func newID() string {
	t := time.Now().UnixNano()
	buf := make([]byte, 8)
	_, _ = rand.Read(buf)
	return fmt.Sprintf("%d%s", t, hex.EncodeToString(buf)[:16])
}
