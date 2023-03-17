package main

import (
	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Post struct {
	ID         primitive.ObjectID `json:"id" bson:"_id"`
	AccountID  string             `json:"account_id" bson:"account_id"`
	Caption    string             `json:"caption"`
	Image      string             `json:"image"`
	Comments   []*Post            `json:"comments"`
	Views      int                `json:"views"`
	Tags       []string           `json:"tags"`
	TotalLikes int                `json:"total_likes" bson:"total_likes"`
	Likes      []string           `json:"likes"`
	RepliesTo  *Post              `json:"replies_to" bson:"replies_to"`
}

func NewPost(accountID, caption, image string, tags []string, repliesTo *Post) *Post {
	return &Post{
		ID:        primitive.NewObjectID(),
		AccountID: accountID,
		Caption:   caption,
		Image:     image,
		Comments:  []*Post{},
		Likes:     []string{},
		Tags:      tags,
		RepliesTo: repliesTo,
	}
}

var validate *validator.Validate

type PostCreationRequest struct {
	Caption   string             `json:"caption" validate:"required"`
	Image     string             `json:"image" validate:"required"`
	Tags      []string           `json:"tags" validate:"required"`
	RepliesTo primitive.ObjectID `json:"replies_to"`
}

func (p *PostCreationRequest) Validate() error {
	return validate.Struct(p)
}
