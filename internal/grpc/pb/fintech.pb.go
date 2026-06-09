package pb

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	_ "github.com/fintech/core/internal/grpc/codec"
)

type Money struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type AccountType int32

const (
	AccountTypeUnspecified AccountType = 0
	AccountTypePersonal    AccountType = 1
	AccountTypeEnterprise  AccountType = 2
)

type AccountStatus int32

const (
	AccountStatusUnspecified AccountStatus = 0
	AccountStatusActive      AccountStatus = 1
	AccountStatusFrozen      AccountStatus = 2
	AccountStatusClosed      AccountStatus = 3
)

type Direction int32

const (
	DirectionUnspecified Direction = 0
	DirectionIn          Direction = 1
	DirectionOut         Direction = 2
)

type TransactionType int32

const (
	TxnTypeUnspecified  TransactionType = 0
	TxnTypeDeposit      TransactionType = 1
	TxnTypeWithdrawal   TransactionType = 2
	TxnTypeTransferIn   TransactionType = 3
	TxnTypeTransferOut  TransactionType = 4
	TxnTypeInterest     TransactionType = 5
)

type TransferStatus int32

const (
	TransferStatusUnspecified TransferStatus = 0
	TransferStatusPending     TransferStatus = 1
	TransferStatusSuccess     TransferStatus = 2
	TransferStatusFailed      TransferStatus = 3
	TransferStatusClearing    TransferStatus = 4
)

type Account struct {
	Id          string        `json:"id"`
	AccountType AccountType   `json:"account_type"`
	AccountNo   string        `json:"account_no"`
	Name        string        `json:"name"`
	IdCardNo    string        `json:"id_card_no"`
	Status      AccountStatus `json:"status"`
	Balance     *Money        `json:"balance"`
	CreatedAt   int64         `json:"created_at"`
	UpdatedAt   int64         `json:"updated_at"`
}

type Transfer struct {
	Id            string         `json:"id"`
	BizId         string         `json:"biz_id"`
	FromAccountId string         `json:"from_account_id"`
	ToAccountId   string         `json:"to_account_id"`
	Amount        *Money         `json:"amount"`
	Remark        string         `json:"remark"`
	Status        TransferStatus `json:"status"`
	IsCrossBank   bool           `json:"is_cross_bank"`
	CreatedAt     int64          `json:"created_at"`
	UpdatedAt     int64          `json:"updated_at"`
}

type LedgerEntry struct {
	Id           string          `json:"id"`
	AccountId    string          `json:"account_id"`
	TransferId   string          `json:"transfer_id"`
	BizId        string          `json:"biz_id"`
	Direction    Direction       `json:"direction"`
	Type         TransactionType `json:"type"`
	Amount       *Money          `json:"amount"`
	BalanceAfter *Money          `json:"balance_after"`
	Remark       string          `json:"remark"`
	CreatedAt    int64           `json:"created_at"`
}

type CreateAccountRequest struct {
	AccountType    AccountType `json:"account_type"`
	Name           string      `json:"name"`
	IdCardNo       string      `json:"id_card_no"`
	Password       string      `json:"password"`
	InitialBalance *Money      `json:"initial_balance"`
}

type GetAccountRequest struct {
	AccountId string `json:"account_id"`
}

type GetBalanceRequest struct {
	AccountId string `json:"account_id"`
}

type BalanceResponse struct {
	AccountId        string `json:"account_id"`
	Balance          *Money `json:"balance"`
	AvailableBalance *Money `json:"available_balance"`
}

type FreezeAccountRequest struct {
	AccountId string `json:"account_id"`
	Reason    string `json:"reason"`
}

type UnfreezeAccountRequest struct {
	AccountId string `json:"account_id"`
}

type CloseAccountRequest struct {
	AccountId string `json:"account_id"`
	Password  string `json:"password"`
}

type ChangePasswordRequest struct {
	AccountId   string `json:"account_id"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type ChangePasswordResponse struct {
	Success bool `json:"success"`
}

type AccountResponse struct {
	Account *Account `json:"account"`
}

type ListAccountsRequest struct {
	Page     int32 `json:"page"`
	PageSize int32 `json:"page_size"`
}

type ListAccountsResponse struct {
	Accounts []*Account `json:"accounts"`
	Total    int32      `json:"total"`
}

type IntraBankTransferRequest struct {
	BizId         string `json:"biz_id"`
	FromAccountId string `json:"from_account_id"`
	ToAccountId   string `json:"to_account_id"`
	Amount        *Money `json:"amount"`
	Remark        string `json:"remark"`
	Password      string `json:"password"`
}

type CrossBankTransferRequest struct {
	BizId         string `json:"biz_id"`
	FromAccountId string `json:"from_account_id"`
	ToBankCode    string `json:"to_bank_code"`
	ToAccountNo   string `json:"to_account_no"`
	ToAccountName string `json:"to_account_name"`
	Amount        *Money `json:"amount"`
	Remark        string `json:"remark"`
	Password      string `json:"password"`
}

type TransferResponse struct {
	Transfer *Transfer `json:"transfer"`
}

type GetTransferRequest struct {
	TransferId string `json:"transfer_id"`
}

type ListTransfersRequest struct {
	AccountId string `json:"account_id"`
	Page      int32  `json:"page"`
	PageSize  int32  `json:"page_size"`
}

type ListTransfersResponse struct {
	Transfers []*Transfer `json:"transfers"`
	Total     int32       `json:"total"`
}

type ListLedgerEntriesRequest struct {
	AccountId string `json:"account_id"`
	Page      int32  `json:"page"`
	PageSize  int32  `json:"page_size"`
}

type ListLedgerEntriesResponse struct {
	Entries []*LedgerEntry `json:"entries"`
	Total   int32          `json:"total"`
}

type ReconcileRequest struct {
	AccountId string `json:"account_id"`
}

type ReconcileResponse struct {
	Consistent         bool   `json:"consistent"`
	BalanceFromAccount *Money `json:"balance_from_account"`
	BalanceFromLedger  *Money `json:"balance_from_ledger"`
	Diff               string `json:"diff"`
}

type UnimplementedAccountServiceServer struct{}

func (UnimplementedAccountServiceServer) CreateAccount(context.Context, *CreateAccountRequest) (*AccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateAccount not implemented")
}
func (UnimplementedAccountServiceServer) GetAccount(context.Context, *GetAccountRequest) (*AccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAccount not implemented")
}
func (UnimplementedAccountServiceServer) GetBalance(context.Context, *GetBalanceRequest) (*BalanceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBalance not implemented")
}
func (UnimplementedAccountServiceServer) FreezeAccount(context.Context, *FreezeAccountRequest) (*AccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FreezeAccount not implemented")
}
func (UnimplementedAccountServiceServer) UnfreezeAccount(context.Context, *UnfreezeAccountRequest) (*AccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnfreezeAccount not implemented")
}
func (UnimplementedAccountServiceServer) CloseAccount(context.Context, *CloseAccountRequest) (*AccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CloseAccount not implemented")
}
func (UnimplementedAccountServiceServer) ChangePassword(context.Context, *ChangePasswordRequest) (*ChangePasswordResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChangePassword not implemented")
}
func (UnimplementedAccountServiceServer) ListAccounts(context.Context, *ListAccountsRequest) (*ListAccountsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAccounts not implemented")
}

type UnimplementedTransferServiceServer struct{}

func (UnimplementedTransferServiceServer) IntraBankTransfer(context.Context, *IntraBankTransferRequest) (*TransferResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IntraBankTransfer not implemented")
}
func (UnimplementedTransferServiceServer) CrossBankTransfer(context.Context, *CrossBankTransferRequest) (*TransferResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CrossBankTransfer not implemented")
}
func (UnimplementedTransferServiceServer) GetTransfer(context.Context, *GetTransferRequest) (*TransferResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTransfer not implemented")
}
func (UnimplementedTransferServiceServer) ListTransfers(context.Context, *ListTransfersRequest) (*ListTransfersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListTransfers not implemented")
}

type UnimplementedLedgerServiceServer struct{}

func (UnimplementedLedgerServiceServer) ListLedgerEntries(context.Context, *ListLedgerEntriesRequest) (*ListLedgerEntriesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListLedgerEntries not implemented")
}
func (UnimplementedLedgerServiceServer) Reconcile(context.Context, *ReconcileRequest) (*ReconcileResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Reconcile not implemented")
}

type AccountServiceServer interface {
	CreateAccount(context.Context, *CreateAccountRequest) (*AccountResponse, error)
	GetAccount(context.Context, *GetAccountRequest) (*AccountResponse, error)
	GetBalance(context.Context, *GetBalanceRequest) (*BalanceResponse, error)
	FreezeAccount(context.Context, *FreezeAccountRequest) (*AccountResponse, error)
	UnfreezeAccount(context.Context, *UnfreezeAccountRequest) (*AccountResponse, error)
	CloseAccount(context.Context, *CloseAccountRequest) (*AccountResponse, error)
	ChangePassword(context.Context, *ChangePasswordRequest) (*ChangePasswordResponse, error)
	ListAccounts(context.Context, *ListAccountsRequest) (*ListAccountsResponse, error)
}

type TransferServiceServer interface {
	IntraBankTransfer(context.Context, *IntraBankTransferRequest) (*TransferResponse, error)
	CrossBankTransfer(context.Context, *CrossBankTransferRequest) (*TransferResponse, error)
	GetTransfer(context.Context, *GetTransferRequest) (*TransferResponse, error)
	ListTransfers(context.Context, *ListTransfersRequest) (*ListTransfersResponse, error)
}

type LedgerServiceServer interface {
	ListLedgerEntries(context.Context, *ListLedgerEntriesRequest) (*ListLedgerEntriesResponse, error)
	Reconcile(context.Context, *ReconcileRequest) (*ReconcileResponse, error)
}

func RegisterAccountServiceServer(s *grpc.Server, srv AccountServiceServer) {
	s.RegisterService(&grpc.ServiceDesc{
		ServiceName: "fintech.AccountService",
		HandlerType: (*AccountServiceServer)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "CreateAccount",
				Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
					in := new(CreateAccountRequest)
					if err := dec(in); err != nil {
						return nil, err
					}
					if interceptor == nil {
						return srv.(AccountServiceServer).CreateAccount(ctx, in)
					}
					info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/fintech.AccountService/CreateAccount"}
					handler := func(ctx context.Context, req interface{}) (interface{}, error) {
						return srv.(AccountServiceServer).CreateAccount(ctx, req.(*CreateAccountRequest))
					}
					return interceptor(ctx, in, info, handler)
				},
			},
			{
				MethodName: "GetAccount",
				Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
					in := new(GetAccountRequest)
					if err := dec(in); err != nil {
						return nil, err
					}
					if interceptor == nil {
						return srv.(AccountServiceServer).GetAccount(ctx, in)
					}
					info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/fintech.AccountService/GetAccount"}
					handler := func(ctx context.Context, req interface{}) (interface{}, error) {
						return srv.(AccountServiceServer).GetAccount(ctx, req.(*GetAccountRequest))
					}
					return interceptor(ctx, in, info, handler)
				},
			},
			{
				MethodName: "GetBalance",
				Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
					in := new(GetBalanceRequest)
					if err := dec(in); err != nil {
						return nil, err
					}
					if interceptor == nil {
						return srv.(AccountServiceServer).GetBalance(ctx, in)
					}
					info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/fintech.AccountService/GetBalance"}
					handler := func(ctx context.Context, req interface{}) (interface{}, error) {
						return srv.(AccountServiceServer).GetBalance(ctx, req.(*GetBalanceRequest))
					}
					return interceptor(ctx, in, info, handler)
				},
			},
			{
				MethodName: "FreezeAccount",
				Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
					in := new(FreezeAccountRequest)
					if err := dec(in); err != nil {
						return nil, err
					}
					if interceptor == nil {
						return srv.(AccountServiceServer).FreezeAccount(ctx, in)
					}
					info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/fintech.AccountService/FreezeAccount"}
					handler := func(ctx context.Context, req interface{}) (interface{}, error) {
						return srv.(AccountServiceServer).FreezeAccount(ctx, req.(*FreezeAccountRequest))
					}
					return interceptor(ctx, in, info, handler)
				},
			},
			{
				MethodName: "UnfreezeAccount",
				Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
					in := new(UnfreezeAccountRequest)
					if err := dec(in); err != nil {
						return nil, err
					}
					if interceptor == nil {
						return srv.(AccountServiceServer).UnfreezeAccount(ctx, in)
					}
					info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/fintech.AccountService/UnfreezeAccount"}
					handler := func(ctx context.Context, req interface{}) (interface{}, error) {
						return srv.(AccountServiceServer).UnfreezeAccount(ctx, req.(*UnfreezeAccountRequest))
					}
					return interceptor(ctx, in, info, handler)
				},
			},
			{
				MethodName: "CloseAccount",
				Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
					in := new(CloseAccountRequest)
					if err := dec(in); err != nil {
						return nil, err
					}
					if interceptor == nil {
						return srv.(AccountServiceServer).CloseAccount(ctx, in)
					}
					info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/fintech.AccountService/CloseAccount"}
					handler := func(ctx context.Context, req interface{}) (interface{}, error) {
						return srv.(AccountServiceServer).CloseAccount(ctx, req.(*CloseAccountRequest))
					}
					return interceptor(ctx, in, info, handler)
				},
			},
			{
				MethodName: "ChangePassword",
				Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
					in := new(ChangePasswordRequest)
					if err := dec(in); err != nil {
						return nil, err
					}
					if interceptor == nil {
						return srv.(AccountServiceServer).ChangePassword(ctx, in)
					}
					info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/fintech.AccountService/ChangePassword"}
					handler := func(ctx context.Context, req interface{}) (interface{}, error) {
						return srv.(AccountServiceServer).ChangePassword(ctx, req.(*ChangePasswordRequest))
					}
					return interceptor(ctx, in, info, handler)
				},
			},
			{
				MethodName: "ListAccounts",
				Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
					in := new(ListAccountsRequest)
					if err := dec(in); err != nil {
						return nil, err
					}
					if interceptor == nil {
						return srv.(AccountServiceServer).ListAccounts(ctx, in)
					}
					info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/fintech.AccountService/ListAccounts"}
					handler := func(ctx context.Context, req interface{}) (interface{}, error) {
						return srv.(AccountServiceServer).ListAccounts(ctx, req.(*ListAccountsRequest))
					}
					return interceptor(ctx, in, info, handler)
				},
			},
		},
		Streams:  []grpc.StreamDesc{},
		Metadata: "proto/fintech.proto",
	}, srv)
}

func RegisterTransferServiceServer(s *grpc.Server, srv TransferServiceServer) {
	s.RegisterService(&grpc.ServiceDesc{
		ServiceName: "fintech.TransferService",
		HandlerType: (*TransferServiceServer)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "IntraBankTransfer",
				Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
					in := new(IntraBankTransferRequest)
					if err := dec(in); err != nil {
						return nil, err
					}
					if interceptor == nil {
						return srv.(TransferServiceServer).IntraBankTransfer(ctx, in)
					}
					info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/fintech.TransferService/IntraBankTransfer"}
					handler := func(ctx context.Context, req interface{}) (interface{}, error) {
						return srv.(TransferServiceServer).IntraBankTransfer(ctx, req.(*IntraBankTransferRequest))
					}
					return interceptor(ctx, in, info, handler)
				},
			},
			{
				MethodName: "CrossBankTransfer",
				Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
					in := new(CrossBankTransferRequest)
					if err := dec(in); err != nil {
						return nil, err
					}
					if interceptor == nil {
						return srv.(TransferServiceServer).CrossBankTransfer(ctx, in)
					}
					info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/fintech.TransferService/CrossBankTransfer"}
					handler := func(ctx context.Context, req interface{}) (interface{}, error) {
						return srv.(TransferServiceServer).CrossBankTransfer(ctx, req.(*CrossBankTransferRequest))
					}
					return interceptor(ctx, in, info, handler)
				},
			},
			{
				MethodName: "GetTransfer",
				Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
					in := new(GetTransferRequest)
					if err := dec(in); err != nil {
						return nil, err
					}
					if interceptor == nil {
						return srv.(TransferServiceServer).GetTransfer(ctx, in)
					}
					info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/fintech.TransferService/GetTransfer"}
					handler := func(ctx context.Context, req interface{}) (interface{}, error) {
						return srv.(TransferServiceServer).GetTransfer(ctx, req.(*GetTransferRequest))
					}
					return interceptor(ctx, in, info, handler)
				},
			},
			{
				MethodName: "ListTransfers",
				Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
					in := new(ListTransfersRequest)
					if err := dec(in); err != nil {
						return nil, err
					}
					if interceptor == nil {
						return srv.(TransferServiceServer).ListTransfers(ctx, in)
					}
					info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/fintech.TransferService/ListTransfers"}
					handler := func(ctx context.Context, req interface{}) (interface{}, error) {
						return srv.(TransferServiceServer).ListTransfers(ctx, req.(*ListTransfersRequest))
					}
					return interceptor(ctx, in, info, handler)
				},
			},
		},
		Streams:  []grpc.StreamDesc{},
		Metadata: "proto/fintech.proto",
	}, srv)
}

func RegisterLedgerServiceServer(s *grpc.Server, srv LedgerServiceServer) {
	s.RegisterService(&grpc.ServiceDesc{
		ServiceName: "fintech.LedgerService",
		HandlerType: (*LedgerServiceServer)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "ListLedgerEntries",
				Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
					in := new(ListLedgerEntriesRequest)
					if err := dec(in); err != nil {
						return nil, err
					}
					if interceptor == nil {
						return srv.(LedgerServiceServer).ListLedgerEntries(ctx, in)
					}
					info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/fintech.LedgerService/ListLedgerEntries"}
					handler := func(ctx context.Context, req interface{}) (interface{}, error) {
						return srv.(LedgerServiceServer).ListLedgerEntries(ctx, req.(*ListLedgerEntriesRequest))
					}
					return interceptor(ctx, in, info, handler)
				},
			},
			{
				MethodName: "Reconcile",
				Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
					in := new(ReconcileRequest)
					if err := dec(in); err != nil {
						return nil, err
					}
					if interceptor == nil {
						return srv.(LedgerServiceServer).Reconcile(ctx, in)
					}
					info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/fintech.LedgerService/Reconcile"}
					handler := func(ctx context.Context, req interface{}) (interface{}, error) {
						return srv.(LedgerServiceServer).Reconcile(ctx, req.(*ReconcileRequest))
					}
					return interceptor(ctx, in, info, handler)
				},
			},
		},
		Streams:  []grpc.StreamDesc{},
		Metadata: "proto/fintech.proto",
	}, srv)
}
