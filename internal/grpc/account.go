package grpcserver

import (
	"context"

	"github.com/fintech/core/internal/account"
	"github.com/fintech/core/internal/grpc/pb"
	"github.com/fintech/core/internal/model"
	fmterrors "github.com/fintech/core/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AccountServer struct {
	pb.UnimplementedAccountServiceServer
	svc *account.Service
}

func NewAccountServer(svc *account.Service) *AccountServer {
	return &AccountServer{svc: svc}
}

func (s *AccountServer) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.AccountResponse, error) {
	initialBalance, err := fromPBMoney(req.InitialBalance)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initial_balance: %v", err)
	}
	var accountType model.AccountType
	switch req.AccountType {
	case pb.AccountTypePersonal:
		accountType = model.AccountTypePersonal
	case pb.AccountTypeEnterprise:
		accountType = model.AccountTypeEnterprise
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid account_type")
	}
	acc, err := s.svc.Create(ctx, account.CreateAccountParams{
		AccountType:    accountType,
		Name:           req.Name,
		IDCardNo:       req.IdCardNo,
		Password:       req.Password,
		InitialBalance: initialBalance,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.AccountResponse{Account: toPBAccount(acc)}, nil
}

func (s *AccountServer) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.AccountResponse, error) {
	acc, err := s.svc.Get(ctx, req.AccountId)
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.AccountResponse{Account: toPBAccount(acc)}, nil
}

func (s *AccountServer) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.BalanceResponse, error) {
	balance, err := s.svc.GetBalance(ctx, req.AccountId)
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.BalanceResponse{
		AccountId:        req.AccountId,
		Balance:          toPBMoney(balance),
		AvailableBalance: toPBMoney(balance),
	}, nil
}

func (s *AccountServer) FreezeAccount(ctx context.Context, req *pb.FreezeAccountRequest) (*pb.AccountResponse, error) {
	err := s.svc.Freeze(ctx, req.AccountId)
	if err != nil {
		return nil, mapError(err)
	}
	acc, err := s.svc.Get(ctx, req.AccountId)
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.AccountResponse{Account: toPBAccount(acc)}, nil
}

func (s *AccountServer) UnfreezeAccount(ctx context.Context, req *pb.UnfreezeAccountRequest) (*pb.AccountResponse, error) {
	err := s.svc.Unfreeze(ctx, req.AccountId)
	if err != nil {
		return nil, mapError(err)
	}
	acc, err := s.svc.Get(ctx, req.AccountId)
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.AccountResponse{Account: toPBAccount(acc)}, nil
}

func (s *AccountServer) CloseAccount(ctx context.Context, req *pb.CloseAccountRequest) (*pb.AccountResponse, error) {
	err := s.svc.Close(ctx, req.AccountId)
	if err != nil {
		return nil, mapError(err)
	}
	acc, err := s.svc.Get(ctx, req.AccountId)
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.AccountResponse{Account: toPBAccount(acc)}, nil
}

func (s *AccountServer) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	err := s.svc.ChangePassword(ctx, req.AccountId, req.OldPassword, req.NewPassword)
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.ChangePasswordResponse{Success: true}, nil
}

func (s *AccountServer) ListAccounts(ctx context.Context, req *pb.ListAccountsRequest) (*pb.ListAccountsResponse, error) {
	limit := int(req.PageSize)
	offset := int(req.Page-1) * int(req.PageSize)
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	accounts, err := s.svc.List(ctx, limit, offset)
	if err != nil {
		return nil, mapError(err)
	}
	pbAccounts := make([]*pb.Account, len(accounts))
	for i, acc := range accounts {
		pbAccounts[i] = toPBAccount(acc)
	}
	return &pb.ListAccountsResponse{
		Accounts: pbAccounts,
		Total:    int32(len(accounts)),
	}, nil
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	var p *fmterrors.Problem
	if fmterrors.As(err, &p) {
		switch p.Status {
		case 400:
			return status.Error(codes.InvalidArgument, p.Detail)
		case 404:
			return status.Error(codes.NotFound, p.Detail)
		case 403:
			return status.Error(codes.PermissionDenied, p.Detail)
		case 409:
			return status.Error(codes.AlreadyExists, p.Detail)
		case 429:
			return status.Error(codes.ResourceExhausted, p.Detail)
		default:
			return status.Error(codes.Internal, p.Detail)
		}
	}
	return status.Error(codes.Internal, err.Error())
}
