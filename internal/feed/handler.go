package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	web "github.com/sina-am/social-media/common"
	"github.com/sina-am/social-media/internal/auth/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *APIServer) CreatePostHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	account := ctx.Value("Account").(*types.Account)

	postReq := &PostCreationRequest{}
	err := json.NewDecoder(r.Body).Decode(postReq)
	if err != nil {
		return err
	}

	if err := postReq.Validate(); err != nil {
		return err
	}

	post := NewPost(account.Id, postReq.Caption, postReq.Image, postReq.Tags, nil)
	err = s.Storage.InsertPost(post)
	if err != nil {
		return err
	}
	return web.WriteJSON(w, http.StatusCreated, map[string]string{"message": "created"})
}

func (s *APIServer) DeletePostHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	postId, err := primitive.ObjectIDFromHex(mux.Vars(r)["id"])
	if err != nil {
		return err
	}

	account := ctx.Value("Account").(*types.Account)
	err = s.Storage.DeleteUserPost(account.Id, postId)
	if err != nil {
		return err
	}

	return web.WriteJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

func (s *APIServer) GetUserPostsHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	account := ctx.Value("Account").(*types.Account)
	posts, err := s.Storage.GetUserPosts(account.Id)
	if err != nil {
		return err
	}

	return web.WriteJSON(w, http.StatusOK, posts)
}

func (s *APIServer) AddCommentHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()
	return web.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"message": "not implemented yes"})
}

func (s *APIServer) AddLikeHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	account := ctx.Value("Account").(*types.Account)
	postId, err := primitive.ObjectIDFromHex(mux.Vars(r)["id"])
	if err != nil {
		return err
	}
	err = s.Storage.InsertLike(postId, account.Id)
	if err != nil {
		return err
	}
	return web.WriteJSON(w, http.StatusOK, map[string]string{"message": "liked!"})
}

func (s *APIServer) GetPosts(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	tagsStr := r.URL.Query().Get("tags")
	if tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		posts, err := s.Storage.GetPostsByTags(tags)
		if err != nil {
			return err
		}

		return web.WriteJSON(w, http.StatusOK, posts)
	}

	accountId := r.URL.Query().Get("account_id")
	if accountId != "" {
		posts, err := s.Storage.GetUserPosts(accountId)
		if err != nil {
			return err
		}

		return web.WriteJSON(w, http.StatusOK, posts)
	}

	return web.WriteJSON(w, http.StatusOK, []string{})
}
