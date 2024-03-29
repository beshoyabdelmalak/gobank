package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type contextKey int

const claimsKey contextKey = iota

type APIServer struct {
	listenAddr string
	store      Storage
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type APIError struct {
	Error string `json:"error"`
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/accounts/{id}", makeHTTPHandleFunc(s.handleGetAccount)).Methods("GET")
	router.HandleFunc("/accounts", makeHTTPHandleFunc(s.handleCreateAccount)).Methods("POST")
	router.HandleFunc("/accounts/{id}", makeHTTPHandleFunc(s.handleDeleteAccount)).Methods("DELETE")
	router.HandleFunc("/login", makeHTTPHandleFunc(s.handleLogin)).Methods("POST")
	router.HandleFunc("/transfer", validateTokenMiddleware(makeHTTPHandleFunc(s.handleTransfer))).Methods("POST")

	log.Println("JSON API server running on port:", s.listenAddr)
	if err := http.ListenAndServe(s.listenAddr, router); err != nil {
		log.Fatal("Could not bring up the server")
	}
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	loginReq := new(LoginRequest)

	if err := json.NewDecoder(r.Body).Decode(loginReq); err != nil {
		return err
	}

	account, err := s.store.GetAccountByIban(loginReq.IBAN)
	if err != nil {
		return err
	}
	if !checkPasswordHash(loginReq.Password, account.EncryptedPassword) {
		return fmt.Errorf("Access Denied")
	}

	token, err := CreateToken(loginReq.IBAN)
	if err != nil {
		return err
	}

	loginResponse := &LoginResponse{
		IBAN:  loginReq.IBAN,
		Token: token,
	}
	return WriteJSON(w, http.StatusOK, loginResponse)
}

func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getId(r)
	if err != nil {
		return err
	}

	account, err := s.store.GetAccountById(id)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createReq := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(createReq); err != nil {
		return err
	}

	account, err := NewAccount(createReq.FirstName, createReq.LastName, createReq.Password)
	if err != nil {
		return err
	}

	if err := s.store.CreateAccount(account); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getId(r)
	if err != nil {
		return err
	}
	err = s.store.DeleteAccount(id)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, map[string]int{"deleted": id})
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	// get the claims of the JWT token
	claims, ok := r.Context().Value(claimsKey).(*Claims)
	if !ok {
		return fmt.Errorf("no claims found in request context")
	}
	fromAccountIban := claims.IBAN

	transferReq := new(TransferRequest)
	if err := json.NewDecoder(r.Body).Decode(&transferReq); err != nil {
		return err
	}

	if err := s.store.TransferFunds(fromAccountIban, transferReq.ToAccountIban, transferReq.Amount); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			_ = WriteJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		}
	}
}

func validateTokenMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			_ = WriteJSON(w, http.StatusBadRequest, APIError{Error: "Authorization header is required"})
			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			_ = WriteJSON(w, http.StatusBadRequest, APIError{Error: "Authorization header must be in the format 'Bearer {token}'"})
			return
		}

		tokenStr := headerParts[1]

		claims, err := ValidateToken(tokenStr)
		if err != nil {
			_ = WriteJSON(w, http.StatusBadRequest, APIError{Error: "Access Denied"})
			return
		}

		// Store the claims in the context
		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func getId(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, fmt.Errorf("Invalid id: %v", idStr)
	}
	return id, nil
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func CreateToken(iban string) (string, error) {
	expirationTime := time.Now().Add(60 * time.Minute)
	jwtKey := os.Getenv("JWT_SECRET")

	claims := &Claims{
		IBAN: iban,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "gobank",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateToken(tokenString string) (*Claims, error) {
	jwtKey := os.Getenv("JWT_SECRET")
	claims := new(Claims)

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
