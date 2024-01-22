package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestHandleCreateAccount(t *testing.T) {
	// Set up the test database
	store := setupTestDB()
	defer tearDownTestDB(store)

	// Create an instance of the APIServer with the test database
	apiServer := NewAPIServer(":8000", store)

	testAccountReq := createTestAccountReq("testFName", "testLName", "testPassword")
	account := createTestAccount(apiServer, t, testAccountReq)
	assert.Equal(t, testAccountReq.FirstName, account.FirstName)
	assert.Equal(t, testAccountReq.LastName, account.LastName)
	assert.NotEmpty(t, account.ID, "Expected non-empty ID")
	assert.NotEmpty(t, account.IBAN, "Expected non-empty IBAN")
	assert.NotEmpty(t, account.Balance, "Expected non-empty Balance")
}

func TestHandleLogin(t *testing.T) {
	// Set up the test database
	store := setupTestDB()
	defer tearDownTestDB(store)

	// Create an instance of the APIServer with the test database
	apiServer := NewAPIServer(":8000", store)

	// create test account
	testAccountReq := createTestAccountReq("testFName", "testLName", "testPassword")
	testAccount := createTestAccount(apiServer, t, testAccountReq)

	loginReq := LoginRequest{
		IBAN:     testAccount.IBAN,
		Password: testAccountReq.Password,
	}
	reqBody, _ := json.Marshal(loginReq)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
	respRec := httptest.NewRecorder()

	handler := http.HandlerFunc(makeHTTPHandleFunc(apiServer.handleLogin))
	handler.ServeHTTP(respRec, req)

	assert.Equal(t, http.StatusOK, respRec.Code)
	var resp LoginResponse
	err := json.Unmarshal(respRec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Token, "Expected non-empty token")
}

func TestHandleLoginNonExistAccount(t *testing.T) {
	// Set up the test database
	store := setupTestDB()
	defer tearDownTestDB(store)

	// Create an instance of the APIServer with the test database
	apiServer := NewAPIServer(":8000", store)

	loginReq := LoginRequest{
		IBAN:     "test_IBAN",
		Password: "test_password",
	}
	reqBody, _ := json.Marshal(loginReq)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
	respRec := httptest.NewRecorder()

	handler := http.HandlerFunc(makeHTTPHandleFunc(apiServer.handleLogin))
	handler.ServeHTTP(respRec, req)

	assert.Equal(t, http.StatusBadRequest, respRec.Code)

	var resp APIError
	err := json.Unmarshal(respRec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Account with IBAN number test_IBAN not found", resp.Error)
}

func TestHandleLoginWrongPassword(t *testing.T) {
	// Set up the test database
	store := setupTestDB()
	defer tearDownTestDB(store)

	// Create an instance of the APIServer with the test database
	apiServer := NewAPIServer(":8000", store)

	// create test account
	testAccountReq := createTestAccountReq("testFName", "testLName", "testPassword")
	testAccount := createTestAccount(apiServer, t, testAccountReq)

	loginReq := LoginRequest{
		IBAN:     testAccount.IBAN,
		Password: "test_password1",
	}
	reqBody, _ := json.Marshal(loginReq)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
	respRec := httptest.NewRecorder()

	handler := http.HandlerFunc(makeHTTPHandleFunc(apiServer.handleLogin))
	handler.ServeHTTP(respRec, req)

	assert.Equal(t, http.StatusBadRequest, respRec.Code)
	var resp APIError
	err := json.Unmarshal(respRec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Access Denied", resp.Error)
}

func TestHandleGetAccount(t *testing.T) {
	// Set up the test database
	store := setupTestDB()
	defer tearDownTestDB(store)

	// Create an instance of the APIServer with the test database
	apiServer := NewAPIServer(":8000", store)

	// create test account
	testAccountReq := createTestAccountReq("testFName", "testLName", "testPassword")
	testAccount := createTestAccount(apiServer, t, testAccountReq)

	req, _ := http.NewRequest("GET", "/accounts", nil)
	req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(testAccount.ID)})
	respRec := httptest.NewRecorder()

	handler := http.HandlerFunc(makeHTTPHandleFunc(apiServer.handleGetAccount))
	handler.ServeHTTP(respRec, req)

	assert.Equal(t, http.StatusOK, respRec.Code)
	var resp Account
	err := json.Unmarshal(respRec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, *testAccount, resp)
}

func TestHandleGetAccountNonExist(t *testing.T) {
	// Set up the test database
	store := setupTestDB()
	defer tearDownTestDB(store)

	// Create an instance of the APIServer with the test database
	apiServer := NewAPIServer(":8000", store)

	req, _ := http.NewRequest("GET", "/accounts", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	respRec := httptest.NewRecorder()

	handler := http.HandlerFunc(makeHTTPHandleFunc(apiServer.handleGetAccount))
	handler.ServeHTTP(respRec, req)

	assert.Equal(t, http.StatusBadRequest, respRec.Code)
	var resp APIError
	err := json.Unmarshal(respRec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Account with id 1 not found", resp.Error)
}

func TestHandleDeleteAccount(t *testing.T) {
	// Set up the test database
	store := setupTestDB()
	defer tearDownTestDB(store)

	// Create an instance of the APIServer with the test database
	apiServer := NewAPIServer(":8000", store)

	// create test account
	testAccountReq := createTestAccountReq("testFName", "testLName", "testPassword")
	testAccount := createTestAccount(apiServer, t, testAccountReq)

	req, _ := http.NewRequest("DELETE", "/accounts", nil)
	req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(testAccount.ID)})
	respRec := httptest.NewRecorder()

	handler := http.HandlerFunc(makeHTTPHandleFunc(apiServer.handleDeleteAccount))
	handler.ServeHTTP(respRec, req)

	assert.Equal(t, http.StatusOK, respRec.Code)
	var resp map[string]int
	err := json.Unmarshal(respRec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, testAccount.ID, resp["deleted"])

	// check account deleted
	_, err = store.GetAccountById(testAccount.ID)
	assert.Equal(t, fmt.Sprintf("Account with id %d not found", testAccount.ID), err.Error())
}

func TestHandleTransfer(t *testing.T) {
	// Set up the test database
	store := setupTestDB()
	defer tearDownTestDB(store)

	// Create an instance of the APIServer with the test database
	apiServer := NewAPIServer(":8000", store)

	// create test accounts
	senderAccountReq := createTestAccountReq("senderFName", "senderLName", "senderPassword")
	senderAccount := createTestAccount(apiServer, t, senderAccountReq)

	receiverAccountReq := createTestAccountReq("receiverFName", "receiverLName", "receiverPassword")
	receiverAccount := createTestAccount(apiServer, t, receiverAccountReq)

	// login with the sender account to get the token
	loginReq := LoginRequest{
		IBAN:     senderAccount.IBAN,
		Password: senderAccountReq.Password,
	}
	reqBody, _ := json.Marshal(loginReq)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
	respRec := httptest.NewRecorder()

	handler := http.HandlerFunc(makeHTTPHandleFunc(apiServer.handleLogin))
	handler.ServeHTTP(respRec, req)

	assert.Equal(t, http.StatusOK, respRec.Code)
	var loginResp LoginResponse
	err := json.Unmarshal(respRec.Body.Bytes(), &loginResp)
	assert.NoError(t, err)
	jwtToken := fmt.Sprintf("Bearer %s", loginResp.Token)

	receiverOldBalance := receiverAccount.Balance
	transferAmount := senderAccount.Balance
	transferRequest := TransferRequest{
		ToAccountIban: receiverAccount.IBAN,
		Amount:        transferAmount,
	}
	reqBody, _ = json.Marshal(transferRequest)

	req, _ = http.NewRequest("POST", "/transfer", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", jwtToken)
	respRec = httptest.NewRecorder()

	handler = http.HandlerFunc(validateTokenMiddleware(makeHTTPHandleFunc(apiServer.handleTransfer)))
	handler.ServeHTTP(respRec, req)

	assert.Equal(t, http.StatusOK, respRec.Code)
	var resp map[string]string
	err = json.Unmarshal(respRec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "success", resp["status"])

	// check updated balances
	senderAccount, _ = store.GetAccountByIban(senderAccount.IBAN)
	receiverAccount, _ = store.GetAccountByIban(receiverAccount.IBAN)
	assert.Equal(t, float64(0), senderAccount.Balance)
	assert.Equal(t, receiverOldBalance+transferAmount, receiverAccount.Balance)
}

func TestHandleTransferInsufficientBalance(t *testing.T) {
	// Set up the test database
	store := setupTestDB()
	defer tearDownTestDB(store)

	// Create an instance of the APIServer with the test database
	apiServer := NewAPIServer(":8000", store)

	// create test accounts
	senderAccountReq := createTestAccountReq("senderFName", "senderLName", "senderPassword")
	senderAccount := createTestAccount(apiServer, t, senderAccountReq)

	receiverAccountReq := createTestAccountReq("receiverFName", "receiverLName", "receiverPassword")
	receiverAccount := createTestAccount(apiServer, t, receiverAccountReq)

	// login with the sender account to get the token
	loginReq := LoginRequest{
		IBAN:     senderAccount.IBAN,
		Password: senderAccountReq.Password,
	}
	reqBody, _ := json.Marshal(loginReq)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
	respRec := httptest.NewRecorder()

	handler := http.HandlerFunc(makeHTTPHandleFunc(apiServer.handleLogin))
	handler.ServeHTTP(respRec, req)

	assert.Equal(t, http.StatusOK, respRec.Code)
	var loginResp LoginResponse
	err := json.Unmarshal(respRec.Body.Bytes(), &loginResp)
	assert.NoError(t, err)
	jwtToken := fmt.Sprintf("Bearer %s", loginResp.Token)

	receiverOldBalance := receiverAccount.Balance
	transferAmount := senderAccount.Balance + 1
	transferRequest := TransferRequest{
		ToAccountIban: receiverAccount.IBAN,
		Amount:        transferAmount,
	}
	reqBody, _ = json.Marshal(transferRequest)

	req, _ = http.NewRequest("POST", "/transfer", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", jwtToken)
	respRec = httptest.NewRecorder()

	handler = http.HandlerFunc(validateTokenMiddleware(makeHTTPHandleFunc(apiServer.handleTransfer)))
	handler.ServeHTTP(respRec, req)

	assert.Equal(t, http.StatusBadRequest, respRec.Code)
	var resp APIError
	err = json.Unmarshal(respRec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Balance not sufficient", resp.Error)

	// check unchanged balances
	senderAccount, _ = store.GetAccountByIban(senderAccount.IBAN)
	receiverAccount, _ = store.GetAccountByIban(receiverAccount.IBAN)
	assert.Equal(t, transferAmount-1, senderAccount.Balance)
	assert.Equal(t, receiverOldBalance, receiverAccount.Balance)
}

func setupTestDB() *PostgresStore {
	store, err := NewPostgresStore()
	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}
	return store
}

func tearDownTestDB(store *PostgresStore) {
	_, err := store.db.Exec("DROP table account")
	if err != nil {
		log.Fatal("Failed to drop test database:", err)
	}
}

func createTestAccountReq(firstName, lastName, password string) *CreateAccountRequest {
	return &CreateAccountRequest{
		FirstName: firstName,
		LastName:  lastName,
		Password:  password,
	}
}

func createTestAccount(apiServer *APIServer, t *testing.T, accReq *CreateAccountRequest) *Account {
	reqBody, _ := json.Marshal(accReq)
	req, _ := http.NewRequest("POST", "/account", bytes.NewBuffer(reqBody))
	respRec := httptest.NewRecorder()

	handler := http.HandlerFunc(makeHTTPHandleFunc(apiServer.handleCreateAccount))
	handler.ServeHTTP(respRec, req)

	var testAccount Account
	err := json.Unmarshal(respRec.Body.Bytes(), &testAccount)
	assert.NoError(t, err)
	return &testAccount
}
