package grpcserver

import (
	"context"

	"github.com/fintech/core/internal/grpc/pb"
	"github.com/fintech/core/internal/ledger"
)

type LedgerServer struct {
	pb.UnimplementedLedgerServiceServer
	svc *ledger.Service
}

func NewLedgerServer(svc *ledger.Service) *LedgerServer {
	return &LedgerServer{svc: svc}
}

func (s *LedgerServer) ListLedgerEntries(ctx context.Context, req *pb.ListLedgerEntriesRequest) (*pb.ListLedgerEntriesResponse, error) {
	limit := int(req.PageSize)
	offset := int(req.Page-1) * int(req.PageSize)
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	entries, err := s.svc.ListByAccount(ctx, req.AccountId, limit, offset)
	if err != nil {
		return nil, mapError(err)
	}
	pbEntries := make([]*pb.LedgerEntry, len(entries))
	for i, e := range entries {
		pbEntries[i] = toPBLedgerEntry(e)
	}
	return &pb.ListLedgerEntriesResponse{
		Entries: pbEntries,
		Total:   int32(len(entries)),
	}, nil
}

func (s *LedgerServer) Reconcile(ctx context.Context, req *pb.ReconcileRequest) (*pb.ReconcileResponse, error) {
	result, err := s.svc.ReconcileAccount(ctx, req.AccountId, "")
	if err != nil {
		return nil, mapError(err)
	}
	isConsistent := result.DiffAmount.Amount.IsZero()
	return &pb.ReconcileResponse{
		Consistent:         isConsistent,
		BalanceFromAccount: toPBMoney(result.BalanceAmount),
		BalanceFromLedger:  toPBMoney(result.LedgerAmount),
		Diff:               result.DiffAmount.Amount.String(),
	}, nil
}
