package main

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID        uuid.UUID `json:"id" sql:"type:uuid"`
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	LastLogin time.Time `json:"last_login"`
	CreatedAt time.Time `json:"created_at"`
	Avatar    string    `json:"avatar"`
	Deleted   bool      `json:"-"`
}

func (a *Account) VerifyPassword(plainPassword string) bool {
	return VerifyPassword(plainPassword, a.Password)
}

type AccountRegisterationRequest struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AccountUpdateRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Avatar   string `json:"avatar"`
	Deleted  bool   `json:"-"`
}

type AccountAuthenticationRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type JWTToken struct {
	Token string `json:"token"`
	Type  string `json:"type"`
}

type FollowRequest struct {
	AccountId uuid.UUID `json:"account_id"`
}

func NewAccount(username, plainPassword, name, email string) (*Account, error) {
	password, err := HashPassword(plainPassword)
	if err != nil {
		return nil, err
	}

	return &Account{
		Username:  username,
		Password:  password,
		Name:      name,
		Email:     email,
		LastLogin: time.Now(),
		CreatedAt: time.Now(),
		Avatar:    "",
		Deleted:   false,
	}, nil
}
