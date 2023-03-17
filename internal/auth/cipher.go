package main

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(plainPassword string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plainPassword), 4)
	return string(bytes), err
}

func VerifyPassword(plainPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(password), []byte(plainPassword))
	return err == nil
}
