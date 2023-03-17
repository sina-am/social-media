package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	web "github.com/sina-am/social-media/common"
	"github.com/sina-am/social-media/internal/auth/client"
	"github.com/sina-am/social-media/internal/auth/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCreatChat(t *testing.T) {
	validate = validator.New()
	accounts := []*types.Account{
		{
			Id:       uuid.NewString(),
			Username: "test1",
			Name:     "test1",
			Email:    "test1@gmail.com",
		},
		{
			Id:       uuid.NewString(),
			Username: "test2",
			Name:     "test2",
			Email:    "test2@gmail.com",
		},
	}
	auth := client.NewFakeGRPCClient(accounts)
	storage := NewMemoryStorage()
	service := NewChatService(storage, auth)
	logger, _ := zap.NewProduction()

	server := APIServer{
		APIServer: web.APIServer{
			Addr:   ":8080",
			Logger: logger,
		},
		Auth:     auth,
		Storage:  storage,
		Service:  service,
		Upgrader: websocket.Upgrader{HandshakeTimeout: 3 * time.Second},
	}
	ctx := context.Background()

	t.Run("/chat unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/chat", nil)
		err := server.createChat(ctx, w, r)
		assert.NotNil(t, err)
		assert.Equal(t, "unauthorized user", err.Error())
	})
	t.Run("/chat method not allowed", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/chat", nil)
		r.Header.Set("Authorization", accounts[0].Id)
		err := server.createChat(ctx, w, r)
		assert.NotNil(t, err)
		assert.Equal(t, "method is not allowed", err.Error())
	})

	t.Run("/chat create", func(t *testing.T) {
		chatIn := &ChatIn{
			Members:   []uuid.UUID{uuid.New(), uuid.New()},
			IsPrivate: true,
		}

		body, err := json.Marshal(chatIn)
		assert.Nil(t, err)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewReader(body))
		r.Header.Set("Authorization", accounts[0].Id)

		err = server.createChat(ctx, w, r)
		assert.Nil(t, err)

		chat := &Chat{}
		assert.Nil(t, json.NewDecoder(w.Body).Decode(chat))
		assert.Equal(t, chatIn.IsPrivate, chat.IsPrivate)
		chatIn.Members = append(chatIn.Members, uuid.MustParse(accounts[0].Id))
		assert.Equal(t, chatIn.Members, chat.Members)
	})

}
func TestServer(t *testing.T) {
	validate = validator.New()
	accounts := []*types.Account{
		{
			Id:       uuid.NewString(),
			Username: "test1",
			Name:     "test1",
			Email:    "test1@gmail.com",
		},
		{
			Id:       uuid.NewString(),
			Username: "test2",
			Name:     "test2",
			Email:    "test2@gmail.com",
		},
	}
	auth := client.NewFakeGRPCClient(accounts)
	storage := NewMemoryStorage()
	service := NewChatService(storage, auth)
	logger, _ := zap.NewProduction()

	server := APIServer{
		APIServer: web.APIServer{
			Addr:   ":8080",
			Logger: logger,
		},
		Auth:     auth,
		Storage:  storage,
		Service:  service,
		Upgrader: websocket.Upgrader{HandshakeTimeout: 3 * time.Second},
	}
	s := httptest.NewServer(http.HandlerFunc(server.wsHandler))
	defer s.Close()
	ctx := context.Background()

	t.Run("UPGRADE /ws with no token", func(t *testing.T) {
		u := "ws" + strings.TrimPrefix(s.URL, "http")
		_, res, err := websocket.DefaultDialer.Dial(u, nil)
		assert.NotNil(t, err)
		assert.Equal(t, res.StatusCode, http.StatusUnauthorized)
	})
	t.Run("UPGRADE /ws with invalid token", func(t *testing.T) {
		u := ("ws" + strings.TrimPrefix(s.URL, "http")) + "?token=" + uuid.NewString()
		_, res, err := websocket.DefaultDialer.Dial(u, nil)
		assert.NotNil(t, err)
		assert.Equal(t, res.StatusCode, http.StatusUnauthorized)
	})

	t.Run("/ws ", func(t *testing.T) {
		u := ("ws" + strings.TrimPrefix(s.URL, "http")) + "?token=" + accounts[0].Id
		ws, _, err := websocket.DefaultDialer.Dial(u, nil)
		assert.Nil(t, err)
		defer ws.Close()

		msg := MessageIn{
			Text: "",
		}
		err = ws.WriteJSON(&msg)
		assert.Nil(t, err)

		response := map[string]string{}
		err = ws.ReadJSON(&response)
		assert.Nil(t, err)
		assert.Equal(t, response["message"], "invalid message received")
	})
	t.Run("/ws invalid uuid", func(t *testing.T) {
		u := ("ws" + strings.TrimPrefix(s.URL, "http")) + "?token=" + accounts[0].Id
		ws, _, err := websocket.DefaultDialer.Dial(u, nil)
		assert.Nil(t, err)
		defer ws.Close()

		msg := map[string]string{
			"to":   "invalid uuid",
			"text": "test message",
		}
		err = ws.WriteJSON(&msg)
		assert.Nil(t, err)

		response := map[string]string{}
		err = ws.ReadJSON(&response)
		assert.Nil(t, err)
		assert.Equal(t, response["message"], "invalid message received")
	})
	t.Run("/ws send to none existent account", func(t *testing.T) {
		u := ("ws" + strings.TrimPrefix(s.URL, "http")) + "?token=" + accounts[0].Id
		ws, _, err := websocket.DefaultDialer.Dial(u, nil)
		assert.Nil(t, err)
		defer ws.Close()

		msg := map[string]string{
			"chat_id": uuid.NewString(),
			"text":    "test message",
		}
		err = ws.WriteJSON(&msg)
		assert.Nil(t, err)

		response := map[string]string{}
		err = ws.ReadJSON(&response)
		assert.Nil(t, err)
		assert.Equal(t, "chat not found", response["message"])
	})
	t.Run("/ws send to none existent chat", func(t *testing.T) {
		u := ("ws" + strings.TrimPrefix(s.URL, "http")) + "?token=" + accounts[0].Id
		ws, _, err := websocket.DefaultDialer.Dial(u, nil)
		assert.Nil(t, err)
		defer ws.Close()

		msg := map[string]string{
			"chat_id": uuid.NewString(),
			"text":    "test message",
		}
		err = ws.WriteJSON(&msg)
		assert.Nil(t, err)

		response := map[string]string{}
		err = ws.ReadJSON(&response)
		assert.Nil(t, err)
		assert.Equal(t, "chat not found", response["message"])
	})
	t.Run("/ws send to online account", func(t *testing.T) {
		u1 := ("ws" + strings.TrimPrefix(s.URL, "http")) + "?token=" + accounts[0].Id
		ws1, _, err := websocket.DefaultDialer.Dial(u1, nil)
		assert.Nil(t, err)
		defer ws1.Close()

		u2 := ("ws" + strings.TrimPrefix(s.URL, "http")) + "?token=" + accounts[1].Id
		ws2, _, err := websocket.DefaultDialer.Dial(u2, nil)
		assert.Nil(t, err)
		defer ws2.Close()

		// Create chat room first
		chatIn := &ChatIn{
			Members:   []uuid.UUID{uuid.MustParse(accounts[0].Id), uuid.MustParse(accounts[1].Id)},
			IsPrivate: true,
		}

		body, err := json.Marshal(chatIn)
		assert.Nil(t, err)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewReader(body))
		r.Header.Set("Authorization", accounts[0].Id)

		err = server.createChat(ctx, w, r)
		assert.Nil(t, err)
		chat := &Chat{}
		assert.Nil(t, json.NewDecoder(w.Body).Decode(chat))

		msg := map[string]string{
			"chat_id": chat.Id.String(),
			"text":    "test message",
		}
		err = ws1.WriteJSON(&msg)
		assert.Nil(t, err)

		var recvMsg Message
		err = ws2.ReadJSON(&recvMsg)
		assert.Nil(t, err)

		assert.Equal(t, recvMsg.Text, "test message")

		response := map[string]string{}
		err = ws1.ReadJSON(&response)
		assert.Nil(t, err)
		assert.Equal(t, "sended", response["message"])

		msg = map[string]string{
			"chat_id": chat.Id.String(),
			"text":    "test message2",
		}
		err = ws2.WriteJSON(&msg)
		assert.Nil(t, err)

		recvMsg2 := Message{}
		err = ws1.ReadJSON(&recvMsg2)
		assert.Nil(t, err)

		assert.Equal(t, recvMsg2.Text, "test message2")

	})
}
