package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	web "github.com/sina-am/social-media/common"
)

func (s *APIServer) getJWTToken(r *http.Request) *JWTToken {
	tokenStr := r.Header.Get("Authorization")
	return &JWTToken{Token: tokenStr, Type: "bearer"}
}

func (s *APIServer) GetAllAccountHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.Storage.GetAllAccount()
	if err != nil {
		return err
	}

	return web.WriteJSON(w, http.StatusOK, accounts)
}

func (s *APIServer) GetUserByIDHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	accountId, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		return err
	}
	account, err := s.Storage.GetByID(accountId)
	if err != nil {
		return err
	}
	return web.WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) GetMyUserHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	jwtToken := s.getJWTToken(r)
	account, err := s.Service.VerifyToken(jwtToken)

	if err != nil {
		return err
	}
	return web.WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) UpdateMyUserHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()
	jwtToken := s.getJWTToken(r)
	accountId, err := s.Service.GetAccountIdFromToken(jwtToken)
	if err != nil {
		return err
	}

	updateReq := &AccountUpdateRequest{}
	err = json.NewDecoder(r.Body).Decode(updateReq)
	if err != nil {
		return err
	}

	err = s.Service.Update(accountId, updateReq)
	if err != nil {
		return err
	}
	return web.WriteJSON(w, http.StatusOK, map[string]string{"message": "account updated"})
}

func (s *APIServer) RegisterHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	accountReq := &AccountRegisterationRequest{}
	err := json.NewDecoder(r.Body).Decode(accountReq)
	if err != nil {
		return err
	}

	newAccount, err := NewAccount(
		accountReq.Username,
		accountReq.Password,
		accountReq.Name,
		accountReq.Email,
	)
	if err != nil {
		return err
	}

	if err := s.Service.Register(newAccount); err != nil {
		return err
	}

	return web.WriteJSON(w, http.StatusOK, newAccount)
}

func (s *APIServer) LoginHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	authReq := &AccountAuthenticationRequest{}
	err := json.NewDecoder(r.Body).Decode(authReq)
	if err != nil {
		return err
	}

	account, err := s.Service.Authenticate(authReq.Username, authReq.Password)
	if err != nil {
		return err
	}

	jwtToken, err := s.Service.ObtainToken(account)
	if err != nil {
		return err
	}

	return web.WriteJSON(w, http.StatusOK, jwtToken)
}

func (s *APIServer) RefreshTokenHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	tokenReq := &JWTToken{}
	err := json.NewDecoder(r.Body).Decode(tokenReq)
	if err != nil {
		return err
	}

	account, err := s.Service.VerifyToken(tokenReq)
	if err != nil {
		return err
	}

	tokenRes, err := s.Service.ObtainToken(account)
	if err != nil {
		return err
	}

	return web.WriteJSON(w, http.StatusOK, tokenRes)
}
func (s *APIServer) GetUserFollowersHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	accountId, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		return err
	}

	accounts, err := s.Storage.GetAccountFollowers(accountId)
	if err != nil {
		return err
	}

	return web.WriteJSON(w, http.StatusOK, accounts)
}
func (s *APIServer) GetMyFollowersHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	myAccountId, err := s.Service.GetAccountIdFromToken(s.getJWTToken(r))
	if err != nil {
		return err
	}

	accounts, err := s.Storage.GetAccountFollowers(myAccountId)
	if err != nil {
		return err
	}

	return web.WriteJSON(w, http.StatusOK, accounts)
}

func (s *APIServer) NewFollowerHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	myAccountId, err := s.Service.GetAccountIdFromToken(s.getJWTToken(r))
	if err != nil {
		return err
	}

	reqBody := &FollowRequest{}
	err = json.NewDecoder(r.Body).Decode(reqBody)
	if err != nil {
		return err
	}

	err = s.Storage.InsertAccountFollower(reqBody.AccountId, myAccountId)
	if err != nil {
		return err
	}

	return web.WriteJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

func (s *APIServer) UnFollowerHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	myAccountId, err := s.Service.GetAccountIdFromToken(s.getJWTToken(r))
	if err != nil {
		return err
	}

	reqBody := &FollowRequest{}
	err = json.NewDecoder(r.Body).Decode(reqBody)
	if err != nil {
		return err
	}

	err = s.Storage.DeleteAccountFollower(reqBody.AccountId, myAccountId)
	if err != nil {
		return err
	}

	return web.WriteJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}
