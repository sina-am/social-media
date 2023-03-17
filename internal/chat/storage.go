package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage interface {
	InsertMessage(ctx context.Context, chatId uuid.UUID, msg *Message) error
	GetChat(ctx context.Context, chatId uuid.UUID) (*Chat, error)
	InsertChat(ctx context.Context, chat *Chat) error
	GetMessages(ctx context.Context, chatId uuid.UUID, count int) ([]*Message, error)
}

type mongoStorage struct {
	client   *mongo.Client
	database string
}

func NewMongoStorage(uri, database string) (*mongoStorage, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	return &mongoStorage{
		client:   client,
		database: database,
	}, nil
}

func (s *mongoStorage) Drop() {
	s.client.Database(s.database).Collection("chats").Drop(context.Background())
}

func (s *mongoStorage) getChatCollection() *mongo.Collection {
	return s.client.Database(s.database).Collection("chats")
}

func (s *mongoStorage) GetChat(ctx context.Context, chatId uuid.UUID) (*Chat, error) {
	res := s.getChatCollection().FindOne(ctx, bson.M{
		"_id": chatId,
	})

	chat := &Chat{}
	if err := res.Decode(chat); err != nil {
		return nil, err
	}
	return chat, nil
}

func (s *mongoStorage) ExistChat(ctx context.Context, members []uuid.UUID) error {
	coll := s.getChatCollection()
	docs, err := coll.Find(ctx, bson.M{
		"members": bson.M{"$eq": members},
	})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil
		}
		return err
	}
	if docs.Next(ctx) {
		return fmt.Errorf("chat with these members already exist")
	}
	return nil
}

func (s *mongoStorage) InsertChat(ctx context.Context, chat *Chat) error {
	if err := s.ExistChat(ctx, chat.Members); err != nil {
		return err
	}

	if chat.Id == uuid.Nil {
		chat.Id = uuid.New()
	}

	_, err := s.getChatCollection().InsertOne(ctx, chat)
	return err
}

func (s *mongoStorage) InsertMessage(ctx context.Context, chatId uuid.UUID, msg *Message) error {
	_, err := s.getChatCollection().UpdateOne(context.TODO(), bson.M{
		"_id": chatId,
	}, bson.M{
		"$push": bson.M{"messages": msg},
	})
	return err
}

func (s *mongoStorage) GetMessages(ctx context.Context, chatId uuid.UUID, count int) ([]*Message, error) {
	return nil, fmt.Errorf("not implemented yet")
}

type memoryStorage struct {
	chats []*Chat
}

func NewMemoryStorage() *memoryStorage {
	return &memoryStorage{
		chats: []*Chat{},
	}
}

func (s *memoryStorage) GetChat(ctx context.Context, chatId uuid.UUID) (*Chat, error) {
	for i := range s.chats {
		if s.chats[i].Id == chatId {
			return s.chats[i], nil
		}
	}
	return nil, fmt.Errorf("chat not found")
}

func (s *memoryStorage) InsertChat(ctx context.Context, chat *Chat) error {
	if chat.Id == uuid.Nil {
		chat.Id, _ = uuid.NewUUID()
	}
	s.chats = append(s.chats, chat)
	return nil
}

func (s *memoryStorage) InsertMessage(ctx context.Context, chatId uuid.UUID, msg *Message) error {
	chat, err := s.GetChat(ctx, chatId)
	if err != nil {
		return err
	}

	chat.Messages = append(chat.Messages, msg)
	return nil
}
func (s *memoryStorage) GetMessages(ctx context.Context, chatId uuid.UUID, count int) ([]*Message, error) {
	chat, err := s.GetChat(ctx, chatId)
	if err != nil {
		return nil, err
	}
	if len(chat.Messages) < count {
		return chat.Messages, nil
	}
	return chat.Messages[:count], nil
}
