package grpcserver

import (
	"context"

	"github.com/fintech/core/internal/grpc/pb"
	"github.com/fintech/core/internal/transfer"
)

type TransferServer struct {
	pb.UnimplementedTransferServiceServer
	svc *transfer.Service
}

func NewTransferServer(svc *transfer.Service) *TransferServer {
	return &TransferServer{svc: svc}
}

func (s *TransferServer) IntraBankTransfer(ctx context.Context, req *pb.IntraBankTransferRequest) (*pb.TransferResponse, error) {
	amount, err := fromPBMoney(req.Amount)
	if err != nil {
		return nil, mapError(err)
	}
	t, err := s.svc.IntraBankTransfer(ctx, transfer.IntraBankTransferParams{
		BizID:         req.BizId,
		FromAccountID: req.FromAccountId,
		ToAccountID:   req.ToAccountId,
		Amount:        amount,
		Remark:        req.Remark,
		Password:      req.Password,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.TransferResponse{Transfer: toPBTransfer(t)}, nil
}

func (s *TransferServer) CrossBankTransfer(ctx context.Context, req *pb.CrossBankTransferRequest) (*pb.TransferResponse, error) {
	amount, err := fromPBMoney(req.Amount)
	if err != nil {
		return nil, mapError(err)
	}
	t, err := s.svc.CrossBankTransfer(ctx, transfer.CrossBankTransferParams{
		BizID:         req.BizId,
		FromAccountID: req.FromAccountId,
		ToBankCode:    req.ToBankCode,
		ToAccountNo:   req.ToAccountNo,
		ToAccountName: req.ToAccountName,
		Amount:        amount,
		Remark:        req.Remark,
		Password:      req.Password,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.TransferResponse{Transfer: toPBTransfer(t)}, nil
}

func (s *TransferServer) GetTransfer(ctx context.Context, req *pb.GetTransferRequest) (*pb.TransferResponse, error) {
	t, err := s.svc.Get(ctx, req.TransferId)
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.TransferResponse{Transfer: toPBTransfer(t)}, nil
}

func (s *TransferServer) ListTransfers(ctx context.Context, req *pb.ListTransfersRequest) (*pb.ListTransfersResponse, error) {
	limit := int(req.PageSize)
	offset := int(req.Page-1) * int(req.PageSize)
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	transfers, err := s.svc.ListByAccount(ctx, req.AccountId, limit, offset)
	if err != nil {
		return nil, mapError(err)
	}
	pbTransfers := make([]*pb.Transfer, len(transfers))
	for i, t := range transfers {
		pbTransfers[i] = toPBTransfer(t)
	}
	return &pb.ListTransfersResponse{
		Transfers: pbTransfers,
		Total:     int32(len(transfers)),
	}, nil
}
