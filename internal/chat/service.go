package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sina-am/social-media/internal/auth/client"
)

type OnlineAccount struct {
	accountID uuid.UUID
	msgCh     chan Message
}

func NewOnlineAccount(accountID uuid.UUID) *OnlineAccount {
	return &OnlineAccount{
		accountID: accountID,
		msgCh:     make(chan Message),
	}
}

func (ws *OnlineAccount) GetAccountID() uuid.UUID {
	return ws.accountID
}

type Service interface {
	AddOnlineAccount(account *OnlineAccount)
	DelOnlineAccount(account *OnlineAccount)
	Deliver(accountId uuid.UUID, msg MessageIn) error
	CreateChat(ctx context.Context, accountId uuid.UUID, chatIn *ChatIn) (*Chat, error)
}

type chatService struct {
	store          Storage
	auth           client.GRPCClient
	onlineAccounts map[uuid.UUID]*OnlineAccount
}

func NewChatService(store Storage, auth client.GRPCClient) *chatService {
	return &chatService{
		store:          store,
		auth:           auth,
		onlineAccounts: make(map[uuid.UUID]*OnlineAccount, 0),
	}
}

func (s *chatService) CreateChat(ctx context.Context, accountId uuid.UUID, chatIn *ChatIn) (*Chat, error) {
	chat := &Chat{
		Id:        uuid.New(),
		Members:   chatIn.Members,
		Messages:  []*Message{},
		IsPrivate: chatIn.IsPrivate,
	}
	chat.Members = append(chat.Members, accountId)

	if err := s.store.InsertChat(ctx, chat); err != nil {
		return nil, err
	}

	return chat, nil
}

func (s *chatService) IsMemberOf(accountId uuid.UUID, chat *Chat) bool {
	for i := range chat.Members {
		if chat.Members[i] == accountId {
			return true
		}
	}
	return false
}
func (s *chatService) Deliver(accountId uuid.UUID, msgIn MessageIn) error {
	chat, err := s.store.GetChat(context.Background(), msgIn.ChatId)
	if err != nil {
		return err
	}

	if !s.IsMemberOf(accountId, chat) {
		return fmt.Errorf("you're not a member of this chat room")
	}

	msg := &Message{
		FromAccountId:  accountId,
		ReplyMessageId: msgIn.ReplyMessageId,
		Text:           msgIn.Text,
		CreatedAt:      time.Now().UTC().Round(time.Second),
	}
	if err := s.store.InsertMessage(context.Background(), chat.Id, msg); err != nil {
		return err
	}

	for _, memberId := range chat.Members {
		// Don't send back the message
		if memberId == accountId {
			continue
		}
		if member, found := s.onlineAccounts[memberId]; found {
			member.msgCh <- *msg
		}
	}
	return nil
}

func (s *chatService) AddOnlineAccount(account *OnlineAccount) {
	s.onlineAccounts[account.GetAccountID()] = account
}

func (s *chatService) DelOnlineAccount(account *OnlineAccount) {
	delete(s.onlineAccounts, account.GetAccountID())
}
