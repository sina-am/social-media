package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	web "github.com/sina-am/social-media/common"
	"github.com/sina-am/social-media/internal/auth/client"
)

type APIServer struct {
	web.APIServer
	Auth    client.GRPCClient
	Storage Storage
	Addr    string
}

func WriteJSON(w http.ResponseWriter, statusCode int, v any) error {
	w.WriteHeader(statusCode)
	w.Header().Add("content-type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

func (s *APIServer) Run() error {
	router := mux.NewRouter()
	router.HandleFunc("/me/posts", s.MakeHTTPHandler(s.AuthenticationMiddleware(s.CreatePostHandler))).Methods("POST")
	router.HandleFunc("/me/posts", s.MakeHTTPHandler(s.AuthenticationMiddleware(s.GetUserPostsHandler))).Methods("GET")
	router.HandleFunc("/me/posts/{id}", s.MakeHTTPHandler(s.AuthenticationMiddleware(s.DeletePostHandler))).Methods("DELETE")

	router.HandleFunc("/posts", s.MakeHTTPHandler(s.AuthenticationMiddleware(s.GetPosts))).Methods("GET")
	router.HandleFunc("/posts/{id}/like", s.MakeHTTPHandler(s.AuthenticationMiddleware(s.AddLikeHandler))).Methods("POST")
	router.HandleFunc("/posts/{id}/comments", s.MakeHTTPHandler(s.AuthenticationMiddleware(s.AddCommentHandler))).Methods("POST")
	return s.APIServer.Run(router)
}
