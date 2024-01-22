package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	GetAccountById(int) (*Account, error)
	GetAccountByIban(string) (*Account, error)
	TransferFunds(fromIban string, toIban string, amount float64) error
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	dbHost := os.Getenv("POSTGRES_HOST")
	dbName := os.Getenv("POSTGRES_DB")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	connStr := fmt.Sprintf("host=%s user=%s dbname=%s password=%s sslmode=disable", dbHost, dbUser, dbName, dbPassword)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init() error {
	return s.createAccountTable()
}

func (s *PostgresStore) createAccountTable() error {
	query := `create table if not exists account (
		id serial primary key,
		first_name varchar(70),
		last_name varchar(70),
		password varchar(100),
		iban varchar(70),
		balance float,
		created_at timestamp
	)`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) CreateAccount(account *Account) error {
	query := `
		insert into account
		(first_name, last_name, password, iban, balance, created_at) 
		values 
		($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	err := s.db.QueryRow(
		query,
		account.FirstName,
		account.LastName,
		account.EncryptedPassword,
		account.IBAN,
		account.Balance,
		account.CreatedAt,
	).Scan(&account.ID)

	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) DeleteAccount(accountId int) error {
	_, err := s.db.Query("delete from account where id=$1", accountId)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) GetAccountById(accountId int) (*Account, error) {
	rows, err := s.db.Query("select * from account where id=$1", accountId)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanAccount(rows)
	}
	return nil, fmt.Errorf("Account with id %d not found", accountId)
}

func (s *PostgresStore) GetAccountByIban(accountIban string) (*Account, error) {
	rows, err := s.db.Query("select * from account where iban=$1", accountIban)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanAccount(rows)
	}
	return nil, fmt.Errorf("Account with IBAN number %s not found", accountIban)
}

func (s *PostgresStore) TransferFunds(fromIban string, toIban string, amount float64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	fromAccount, err := s.lockAccount(tx, fromIban)
	if err != nil {
		return err
	}
	if amount > fromAccount.Balance {
		return fmt.Errorf("Balance not sufficient")
	}

	updateBalance := func(iban string, balance float64) error {
		_, err := tx.Exec("update account set balance = $2 where iban = $1", iban, balance)
		return err
	}

	if err := updateBalance(fromIban, fromAccount.Balance-amount); err != nil {
		return err
	}

	toAccount, err := s.lockAccount(tx, toIban)
	if err != nil {
		return err
	}

	if err := updateBalance(toIban, toAccount.Balance+amount); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *PostgresStore) lockAccount(tx *sql.Tx, iban string) (*Account, error) {
	// lockAccount locks the specified account for update and returns its details
	var account Account

	query := `SELECT iban, balance FROM account WHERE iban = $1 FOR UPDATE;`
	err := tx.QueryRow(query, iban).Scan(&account.IBAN, &account.Balance)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func scanAccount(rows *sql.Rows) (*Account, error) {
	account := new(Account)
	err := rows.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.EncryptedPassword,
		&account.IBAN,
		&account.Balance,
		&account.CreatedAt,
	)
	return account, err
}
