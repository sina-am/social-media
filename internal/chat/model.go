package main

import (
	"time"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
)

var validate *validator.Validate

type Chat struct {
	Id        uuid.UUID   `json:"id" bson:"_id"`
	Members   []uuid.UUID `json:"members" bson:"members"`
	Messages  []*Message  `json:"messages" bson:"messages"`
	IsPrivate bool        `json:"is_private" bson:"is_private"`
}

type ChatIn struct {
	Members   []uuid.UUID `json:"members" validate:"required"`
	IsPrivate bool        `json:"is_private" validate:"required"`
}

func (in *ChatIn) Validate() error {
	return validate.Struct(in)
}

type Message struct {
	FromAccountId  uuid.UUID `json:"from" bson:"from_account_id" validate:"required"`
	ReplyMessageId uuid.UUID `json:"reply_to" bson:"replay_message_id,omitempty"`
	Text           string    `json:"text" bson:"text" validate:"required"`
	CreatedAt      time.Time `json:"created_at" bson:"created_at" validate:"required"`
}

type MessageIn struct {
	ChatId         uuid.UUID `json:"chat_id" validate:"required"`
	ReplyMessageId uuid.UUID `json:"reply_to" `
	Text           string    `json:"text"`
}

func (in *MessageIn) Validate() error {
	return validate.Struct(in)
}
