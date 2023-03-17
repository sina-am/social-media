package main

import (
	"net/http"

	"github.com/gorilla/mux"
	web "github.com/sina-am/social-media/common"
)

type APIServer struct {
	web.APIServer
	Service AuthService
	Storage Storage
	Addr    string
	Router  *mux.Router
}

func (s *APIServer) Run() error {
	s.Router.HandleFunc("/accounts", s.MakeHTTPHandler(s.RegisterHandler)).Methods("POST")
	s.Router.HandleFunc("/accounts", s.MakeHTTPHandler(s.GetUserByIDHandler)).Methods("GET").Queries("id", "{id}")
	s.Router.HandleFunc("/accounts", s.MakeHTTPHandler(s.GetAllAccountHandler)).Methods("GET")

	s.Router.HandleFunc("/accounts/me/followers", s.MakeHTTPHandler(s.GetMyFollowersHandler)).Methods("GET")
	s.Router.HandleFunc("/accounts/{id}/followers", s.MakeHTTPHandler(s.GetUserFollowersHandler)).Methods("GET")

	s.Router.HandleFunc("/accounts/me", s.MakeHTTPHandler(s.GetMyUserHandler)).Methods("GET")
	s.Router.HandleFunc("/accounts/me", s.MakeHTTPHandler(s.UpdateMyUserHandler)).Methods(http.MethodPut)
	s.Router.HandleFunc("/accounts/me/followers", s.MakeHTTPHandler(s.GetMyFollowersHandler)).Methods("GET")
	s.Router.HandleFunc("/accounts/follow", s.MakeHTTPHandler(s.NewFollowerHandler)).Methods("POST")
	s.Router.HandleFunc("/accounts/follow", s.MakeHTTPHandler(s.UnFollowerHandler)).Methods("DELETE")

	s.Router.HandleFunc("/obtain", s.MakeHTTPHandler(s.LoginHandler)).Methods("POST")
	s.Router.HandleFunc("/refresh", s.MakeHTTPHandler(s.RefreshTokenHandler)).Methods("POST")
	return s.APIServer.Run(s.Router)
}
