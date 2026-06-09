package grpcserver

import (
	"net"

	"github.com/fintech/core/internal/account"
	"github.com/fintech/core/internal/grpc/pb"
	"github.com/fintech/core/internal/ledger"
	"github.com/fintech/core/internal/transfer"
	"google.golang.org/grpc"
)

type Server struct {
	grpcSrv *grpc.Server
	addr    string
}

func NewServer(addr string, accountSvc *account.Service, transferSvc *transfer.Service, ledgerSvc *ledger.Service) *Server {
	s := grpc.NewServer()

	pb.RegisterAccountServiceServer(s, NewAccountServer(accountSvc))
	pb.RegisterTransferServiceServer(s, NewTransferServer(transferSvc))
	pb.RegisterLedgerServiceServer(s, NewLedgerServer(ledgerSvc))

	return &Server{
		grpcSrv: s,
		addr:    addr,
	}
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	return s.grpcSrv.Serve(lis)
}

func (s *Server) Stop() {
	s.grpcSrv.GracefulStop()
}
