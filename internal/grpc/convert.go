package grpcserver

import (
	"github.com/fintech/core/internal/grpc/pb"
	"github.com/fintech/core/internal/model"
	"github.com/fintech/core/pkg/money"
)

func toPBMoney(m money.Money) *pb.Money {
	return &pb.Money{
		Amount:   m.Amount.String(),
		Currency: m.Currency,
	}
}

func fromPBMoney(m *pb.Money) (money.Money, error) {
	if m == nil {
		return money.Zero(), nil
	}
	return money.NewFromString(m.Amount)
}

func toPBAccount(a *model.Account) *pb.Account {
	if a == nil {
		return nil
	}
	var status pb.AccountStatus
	switch a.Status {
	case model.AccountStatusActive:
		status = pb.AccountStatusActive
	case model.AccountStatusFrozen:
		status = pb.AccountStatusFrozen
	case model.AccountStatusClosed:
		status = pb.AccountStatusClosed
	default:
		status = pb.AccountStatusUnspecified
	}
	var accountType pb.AccountType
	switch a.AccountType {
	case model.AccountTypePersonal:
		accountType = pb.AccountTypePersonal
	case model.AccountTypeEnterprise:
		accountType = pb.AccountTypeEnterprise
	default:
		accountType = pb.AccountTypeUnspecified
	}
	return &pb.Account{
		Id:          a.ID,
		AccountType: accountType,
		AccountNo:   a.AccountNo,
		Name:        a.Name,
		IdCardNo:    a.IDCardNo,
		Status:      status,
		Balance:     toPBMoney(a.Balance),
		CreatedAt:   a.CreatedAt.Unix(),
		UpdatedAt:   a.UpdatedAt.Unix(),
	}
}

func toPBTransfer(t *model.Transfer) *pb.Transfer {
	if t == nil {
		return nil
	}
	var status pb.TransferStatus
	switch t.Status {
	case model.TransferStatusPending:
		status = pb.TransferStatusPending
	case model.TransferStatusSuccess:
		status = pb.TransferStatusSuccess
	case model.TransferStatusFailed:
		status = pb.TransferStatusFailed
	case model.TransferStatusSettling:
		status = pb.TransferStatusClearing
	default:
		status = pb.TransferStatusUnspecified
	}
	return &pb.Transfer{
		Id:            t.ID,
		BizId:         t.BizID,
		FromAccountId: t.FromAccountID,
		ToAccountId:   t.ToAccountID,
		Amount:        toPBMoney(t.Amount),
		Remark:        t.Remark,
		Status:        status,
		IsCrossBank:   t.IsCrossBank,
		CreatedAt:     t.CreatedAt.Unix(),
		UpdatedAt:     t.UpdatedAt.Unix(),
	}
}

func toPBLedgerEntry(e *model.LedgerEntry) *pb.LedgerEntry {
	if e == nil {
		return nil
	}
	var direction pb.Direction
	switch e.Direction {
	case model.DirectionIn:
		direction = pb.DirectionIn
	case model.DirectionOut:
		direction = pb.DirectionOut
	default:
		direction = pb.DirectionUnspecified
	}
	var txnType pb.TransactionType
	switch e.Type {
	case model.TxnTypeDeposit:
		txnType = pb.TxnTypeDeposit
	case model.TxnTypeWithdraw:
		txnType = pb.TxnTypeWithdrawal
	case model.TxnTypeTransferIn:
		txnType = pb.TxnTypeTransferIn
	case model.TxnTypeTransferOut:
		txnType = pb.TxnTypeTransferOut
	case model.TxnTypeInterest:
		txnType = pb.TxnTypeInterest
	default:
		txnType = pb.TxnTypeUnspecified
	}
	return &pb.LedgerEntry{
		Id:           e.ID,
		AccountId:    e.AccountID,
		TransferId:   e.TransferID,
		BizId:        e.BizID,
		Direction:    direction,
		Type:         txnType,
		Amount:       toPBMoney(e.Amount),
		BalanceAfter: toPBMoney(e.BalanceAfter),
		Remark:       e.Remark,
		CreatedAt:    e.CreatedAt.Unix(),
	}
}
