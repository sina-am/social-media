package main

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMongoStorage(t *testing.T) {
	storage, err := NewMongoStorage("mongodb://localhost", "chats_test")
	assert.Nil(t, err)
	defer storage.Drop()

	ctx := context.Background()
	t.Run("test creating chat", func(t *testing.T) {
		chat := &Chat{
			Id:        uuid.New(),
			Members:   []uuid.UUID{uuid.New(), uuid.New()},
			Messages:  []*Message{},
			IsPrivate: true,
		}
		assert.Nil(t, storage.InsertChat(ctx, chat))
	})
	t.Run("test recreating chat", func(t *testing.T) {
		accountId1 := uuid.New()
		accountId2 := uuid.New()
		chat := &Chat{
			Id:        uuid.New(),
			Members:   []uuid.UUID{accountId1, accountId2},
			Messages:  []*Message{},
			IsPrivate: true,
		}
		assert.Nil(t, storage.InsertChat(ctx, chat))

		assert.NotNil(t, storage.InsertChat(ctx, chat))
	})
	t.Run("test creating chat with more members", func(t *testing.T) {
		accountId1 := uuid.New()
		accountId2 := uuid.New()
		chat := &Chat{
			Id:        uuid.New(),
			Members:   []uuid.UUID{accountId1, accountId2, uuid.New()},
			Messages:  []*Message{},
			IsPrivate: true,
		}
		assert.Nil(t, storage.InsertChat(ctx, chat))

		chat.Id = uuid.New()
		chat.Members = []uuid.UUID{accountId1, accountId2}
		assert.Nil(t, storage.InsertChat(ctx, chat))
	})
	t.Run("test get chat", func(t *testing.T) {
		chat := &Chat{
			Id:        uuid.New(),
			Members:   []uuid.UUID{uuid.New(), uuid.New()},
			Messages:  []*Message{},
			IsPrivate: true,
		}
		storage.InsertChat(ctx, chat)
		chatDb, err := storage.GetChat(ctx, chat.Id)
		assert.Nil(t, err)
		assert.Equal(t, chat, chatDb)
	})
	t.Run("test insert message", func(t *testing.T) {
		chat := &Chat{
			Id:        uuid.New(),
			Members:   []uuid.UUID{uuid.New(), uuid.New()},
			Messages:  []*Message{},
			IsPrivate: true,
		}
		storage.InsertChat(ctx, chat)

		msg := &Message{
			FromAccountId: uuid.New(),
			Text:          "test message",
			CreatedAt:     time.Now().UTC().Round(time.Second),
		}
		err := storage.InsertMessage(ctx, chat.Id, msg)
		assert.Nil(t, err)

		chatDb, err := storage.GetChat(ctx, chat.Id)
		assert.Nil(t, err)
		assert.Equal(t, msg, chatDb.Messages[0])
	})
}
