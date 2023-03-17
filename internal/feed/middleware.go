package main

import (
	"context"
	"fmt"
	"net/http"

	web "github.com/sina-am/social-media/common"
	"github.com/sina-am/social-media/internal/auth/types"
)

func (s *APIServer) AuthenticationMiddleware(f web.APIFunc) web.APIFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		token := r.Header.Get("Authorization")
		if token == "" {
			return fmt.Errorf("unauthorized user")
		}

		account, err := s.Auth.ObtainAccountRPC(
			ctx,
			&types.JWTToken{Token: token, Type: "bearer"},
		)
		if err != nil {
			return err
		}

		ctx = context.WithValue(ctx, "Account", account)
		return f(ctx, w, r)
	}
}
