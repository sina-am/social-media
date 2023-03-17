package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	web "github.com/sina-am/social-media/common"
	"github.com/sina-am/social-media/internal/auth/client"
	"github.com/sina-am/social-media/internal/auth/types"
)

type APIServer struct {
	web.APIServer
	Auth     client.GRPCClient
	Upgrader websocket.Upgrader
	Service  Service
	Storage  Storage
}

func (s *APIServer) createChat(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	token := r.Header.Get("Authorization")
	if token == "" {
		return web.Errorf(http.StatusUnauthorized, "unauthorized user")
	}

	account, err := s.Auth.ObtainAccountRPC(
		ctx,
		&types.JWTToken{Token: token, Type: "bearer"},
	)
	if err != nil {
		return err
	}

	if r.Method != http.MethodPost {
		return web.Errorf(http.StatusMethodNotAllowed, "method is not allowed")
	}

	defer r.Body.Close()

	chatIn := &ChatIn{}
	if err := json.NewDecoder(r.Body).Decode(chatIn); err != nil {
		return web.Errorf(http.StatusBadRequest, err.Error())
	}

	chat, err := s.Service.CreateChat(ctx, uuid.MustParse(account.Id), chatIn)
	if err != nil {
		return web.Errorf(http.StatusBadRequest, err.Error())
	}

	return web.WriteJSON(w, http.StatusCreated, chat)
}

func (s *APIServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	token := &types.JWTToken{
		Token: r.URL.Query().Get("token"),
		Type:  "bearer",
	}
	account, err := s.Auth.ObtainAccountRPC(ctx, token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("invalid token"))
		return
	}

	conn, err := s.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	accountId := uuid.MustParse(account.Id)
	onlineAccount := NewOnlineAccount(accountId)
	s.Service.AddOnlineAccount(onlineAccount)
	defer func() {
		s.Service.DelOnlineAccount(onlineAccount)
		conn.Close()
	}()

	go func(ctx context.Context) {
		for msg := range onlineAccount.msgCh {
			conn.WriteJSON(msg)
		}
	}(ctx)

	for {
		var msgIn MessageIn
		if err := conn.ReadJSON(&msgIn); err != nil {
			conn.WriteJSON(web.Errorf(http.StatusBadRequest, "invalid message received"))
			continue
		}
		if err := msgIn.Validate(); err != nil {
			conn.WriteJSON(web.Errorf(http.StatusBadRequest, "invalid message received"))
			continue
		}

		if err := s.Service.Deliver(accountId, msgIn); err != nil {
			conn.WriteJSON(web.Errorf(http.StatusBadRequest, err.Error()))
			continue
		}
		conn.WriteJSON(&map[string]string{"message": "sended"})
	}
}

func (s *APIServer) Run() error {
	router := mux.NewRouter()
	router.HandleFunc("/chat", s.MakeHTTPHandler(s.createChat)).Methods("POST")
	router.HandleFunc("/ws", s.wsHandler)
	return s.APIServer.Run(router)
}
