package main

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage interface {
	InsertPost(*Post) error
	DeletePost(postId primitive.ObjectID) error
	InsertLike(postId primitive.ObjectID, accountId string) error

	DeleteUserPost(accountId string, postId primitive.ObjectID) error
	GetUserPosts(accountId string) ([]*Post, error)

	GetPostsByTags(tags []string) ([]*Post, error)
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

func (s *mongoStorage) InsertPost(p *Post) error {
	_, err := s.getPostCollection().InsertOne(context.TODO(), p)
	return err
}

func (s *mongoStorage) DeletePost(postId primitive.ObjectID) error {
	_, err := s.getPostCollection().DeleteOne(context.TODO(), bson.M{"_id": postId})
	return err
}

func (s *mongoStorage) InsertLike(postId primitive.ObjectID, accountId string) error {
	_, err := s.getPostCollection().UpdateByID(
		context.TODO(), postId,
		bson.M{
			"$push": bson.M{"likes": accountId},
			"$inc":  bson.M{"total_likes": 1},
		},
	)

	/*
		{
			"$set": {"$push": {"likes": accountID}},
			"inc": {"total_likes": 1}
		}
	*/
	return err
}

func (s *mongoStorage) DeleteUserPost(accountId string, postId primitive.ObjectID) error {
	_, err := s.getPostCollection().DeleteOne(
		context.TODO(),
		bson.M{
			"$and": []bson.M{
				{"_id": postId},
				{"account_id": accountId},
			},
		},
	)

	return err
}

func (s *mongoStorage) GetPostsByTags(tags []string) ([]*Post, error) {
	coll := s.getPostCollection()
	cur, err := coll.Find(
		context.TODO(),
		bson.M{"tags": bson.M{"$all": tags}},
	)
	if err != nil {
		return nil, err
	}

	return s.fetchPosts(cur)
}

func (s *mongoStorage) GetUserPosts(AccountID string) ([]*Post, error) {
	cur, err := s.getPostCollection().Find(
		context.TODO(),
		bson.M{"account_id": AccountID},
	)
	if err != nil {
		return nil, err
	}

	return s.fetchPosts(cur)
}

func (s *mongoStorage) fetchPosts(cur *mongo.Cursor) ([]*Post, error) {
	posts := []*Post{}
	for cur.Next(context.TODO()) {
		post := &Post{}
		if err := cur.Decode(post); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (s *mongoStorage) getPostCollection() *mongo.Collection {
	return s.client.Database(s.database).Collection("posts")
}

type memoryStorage struct {
	posts []*Post
}

func NewMemoryStorage() (*memoryStorage, error) {
	return &memoryStorage{
		posts: make([]*Post, 10),
	}, nil
}

func (s *memoryStorage) InsertPost(p *Post) error {
	s.posts = append(s.posts, p)
	return nil
}

func (s *memoryStorage) GetUserPosts(AccountID string) ([]*Post, error) {
	myPosts := []*Post{}
	for _, post := range s.posts {
		if post.AccountID == AccountID {
			myPosts = append(myPosts, post)
		}
	}
	return myPosts, nil
}
