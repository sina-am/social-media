package client

import (
	"context"
	"fmt"
	"time"

	"github.com/sina-am/social-media/internal/auth/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClient interface {
	ObtainAccountRPC(ctx context.Context, jwtToken *types.JWTToken) (*types.Account, error)
	GetAccountByIdRPC(ctx context.Context, accountId string) (*types.Account, error)
}

type gRPCClient struct {
	Addr string
}

func NewGRPCClient(addr string) *gRPCClient {
	return &gRPCClient{
		Addr: addr,
	}
}

func (c *gRPCClient) ObtainAccountRPC(ctx context.Context, jwtToken *types.JWTToken) (*types.Account, error) {
	conn, err := grpc.Dial(c.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	authClient := types.NewAuthenticationClient(conn)

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	account, err := authClient.ObtainAccount(ctx, jwtToken)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (c *gRPCClient) GetAccountByIdRPC(ctx context.Context, accountId string) (*types.Account, error) {
	conn, err := grpc.Dial(c.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	authClient := types.NewAuthenticationClient(conn)

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	account, err := authClient.GetAccountByID(ctx, &types.GetAccountRequest{
		AccountId: accountId,
	})
	if err != nil {
		return nil, err
	}
	return account, nil
}

type fakeGRPCClient struct {
	accounts []*types.Account
}

func NewFakeGRPCClient(accounts []*types.Account) *fakeGRPCClient {
	return &fakeGRPCClient{
		accounts,
	}
}

func (c *fakeGRPCClient) ObtainAccountRPC(ctx context.Context, jwtToken *types.JWTToken) (*types.Account, error) {
	if jwtToken.Token == "" {
		return nil, fmt.Errorf("invalid token")
	}

	return c.GetAccountByIdRPC(ctx, jwtToken.Token)
}

func (c *fakeGRPCClient) GetAccountByIdRPC(ctx context.Context, accountId string) (*types.Account, error) {
	for _, account := range c.accounts {
		if account.Id == accountId {
			return account, nil
		}
	}
	return nil, fmt.Errorf("invalid account id")
}
