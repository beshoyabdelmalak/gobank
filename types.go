package main

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	IBAN     string `json:"iban"`
	Password string `json:"password"`
}

type LoginResponse struct {
	IBAN  string `json:"iban"`
	Token string `json:"token"`
}

type CreateAccountRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Password  string `json:"password"`
}

type Claims struct {
	IBAN string `json:"iban"`
	jwt.StandardClaims
}

type TransferRequest struct {
	ToAccountIban string  `json:"toAccountIban"`
	Amount        float64 `json:"amount"`
}

type Account struct {
	ID                int       `json:"id"`
	FirstName         string    `json:"firstName"`
	LastName          string    `json:"lastName"`
	EncryptedPassword string    `json:"encryptedPassword"`
	IBAN              string    `json:"iban"`
	Balance           float64   `json:"balance"`
	CreatedAt         time.Time `json:"createdAt"`
}

func NewAccount(firstName, lastName, password string) (*Account, error) {
	encryptedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	return &Account{
		ID:                rand.Intn(10000),
		FirstName:         firstName,
		LastName:          lastName,
		EncryptedPassword: encryptedPassword,
		IBAN:              strconv.Itoa(rand.Intn(1000000)),
		Balance:           float64(rand.Intn(1000000)),
		CreatedAt:         time.Now().UTC(),
	}, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
