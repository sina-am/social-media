package main

import (
	"context"
	"log"
	"net"

	"github.com/google/uuid"
	"github.com/sina-am/social-media/internal/auth/types"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	types.UnimplementedAuthenticationServer
	Service AuthService
	Addr    string
}

func (s *GRPCServer) ObtainAccount(ctx context.Context, in *types.JWTToken) (*types.Account, error) {
	log.Printf("Received: %v", in.GetToken())
	token := &JWTToken{
		Token: in.GetToken(),
		Type:  in.GetType(),
	}
	account, err := s.Service.VerifyToken(token)
	if err != nil {
		return nil, err
	}
	log.Printf("token verified with account %s", account.ID)
	return &types.Account{
		Id:       string(account.ID.String()),
		Username: account.Username,
		Name:     account.Name,
		Email:    account.Email,
	}, nil
}
func (s *GRPCServer) GetAccountByID(ctx context.Context, in *types.GetAccountRequest) (*types.Account, error) {
	accountId, err := uuid.Parse(in.AccountId)
	if err != nil {
		return nil, err
	}
	account, err := s.Service.GetAccountByID(accountId)
	if err != nil {
		return nil, err
	}

	return &types.Account{
		Id:       string(account.ID.String()),
		Username: account.Username,
		Name:     account.Name,
		Email:    account.Email,
	}, nil
}

func (s *GRPCServer) Run() error {
	listen, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	types.RegisterAuthenticationServer(grpcServer, s)
	return grpcServer.Serve(listen)
}
