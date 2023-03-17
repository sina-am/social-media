package main

import (
	"database/sql"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Storage interface {
	InsertAccount(*Account) error
	GetAllAccount() ([]*Account, error)
	GetByUsername(username string) (*Account, error)
	GetByID(id uuid.UUID) (*Account, error)
	Update(*Account) error
	InsertAccountFollower(accountId uuid.UUID, followerId uuid.UUID) error
	DeleteAccountFollower(accountId uuid.UUID, followerId uuid.UUID) error
	GetAccountFollowers(uuid.UUID) ([]*Account, error)
}

type postgresStorage struct {
	db *sql.DB
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,
	)

	if err != nil {
		return err
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS accounts (
			id uuid DEFAULT uuid_generate_v4 (),
			username VARCHAR(255) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			last_login TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL,
			avatar VARCHAR(512),
			deleted boolean DEFAULT 'f',
			
			PRIMARY KEY (id)
		);
		CREATE TABLE IF NOT EXISTS followers (
			account_id uuid NOT NULL,
			follower_id uuid NOT NULL,

			FOREIGN KEY (account_id) REFERENCES accounts (id)
				ON DELETE CASCADE,
			FOREIGN KEY (follower_id) REFERENCES accounts (id)
				ON DELETE CASCADE,

			PRIMARY KEY (account_id, follower_id)
		);
	`)
	if err != nil {
		return err
	}
	return nil
}

func NewPostgresStorage(connStr string) (*postgresStorage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	err = migrate(db)
	if err != nil {
		return nil, err
	}

	return &postgresStorage{
		db: db,
	}, nil
}

func (s *postgresStorage) InsertAccount(account *Account) error {
	query := `
		INSERT INTO 
			accounts(
				username, password, name, 
				email, last_login, created_at, 
				avatar
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7);
	`

	_, err := s.db.Exec(
		query, account.Username,
		account.Password, account.Name,
		account.Email, account.LastLogin,
		account.CreatedAt, account.Avatar,
	)

	return err
}
func (s *postgresStorage) GetAllAccount() ([]*Account, error) {
	query := `
		SELECT 
			id, username, password, name,
			email, last_login, created_at,
			avatar
		FROM accounts
		WHERE deleted = false;
	`

	result, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}

	accounts := []*Account{}

	for result.Next() {
		account := &Account{}
		err := result.Scan(
			&account.ID,
			&account.Username,
			&account.Password,
			&account.Name,
			&account.Email,
			&account.LastLogin,
			&account.CreatedAt,
			&account.Avatar,
		)

		if err != nil {
			return nil, err
		}

		account.Deleted = false
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func (s *postgresStorage) GetByUsername(username string) (*Account, error) {
	query := `
		SELECT
			id, username, password, name,
			email, last_login, created_at,
			avatar
		FROM accounts 
		WHERE deleted = false AND username = $1
	`

	account := &Account{}
	err := s.db.QueryRow(query, username).Scan(
		&account.ID,
		&account.Username,
		&account.Password,
		&account.Name,
		&account.Email,
		&account.LastLogin,
		&account.CreatedAt,
		&account.Avatar,
	)
	if err != nil {
		return nil, err
	}

	return account, nil
}
func (s *postgresStorage) GetByID(id uuid.UUID) (*Account, error) {
	query := `
		SELECT
			id, username, password, name,
			email, last_login, created_at,
			avatar
		FROM accounts 
		WHERE deleted = false AND id = $1
	`

	account := &Account{}
	err := s.db.QueryRow(query, id).Scan(
		&account.ID,
		&account.Username,
		&account.Password,
		&account.Name,
		&account.Email,
		&account.LastLogin,
		&account.CreatedAt,
		&account.Avatar,
	)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (s *postgresStorage) Update(account *Account) error {
	query := `
		UPDATE accounts
		SET 
			name=$1,
			email=$2,
			password=$3,
			last_login=$4,
			avatar=$5,
			deleted=$6
		WHERE id=$7 AND deleted  = false;
	`

	_, err := s.db.Exec(
		query,
		account.Name,
		account.Email,
		account.Password,
		account.LastLogin,
		account.Avatar,
		account.Deleted,
		account.ID.String(),
	)

	return err
}

func (s *postgresStorage) InsertAccountFollower(accountId, followerId uuid.UUID) error {
	query := `
		INSERT INTO followers(account_id, follower_id)
		VALUES ($1, $2)	
	`

	_, err := s.db.Exec(query, accountId.String(), followerId.String())
	return err
}

func (s *postgresStorage) DeleteAccountFollower(accountId uuid.UUID, followerId uuid.UUID) error {
	query := `
		DELETE FROM followers
		WHERE (account_id = $1 AND follower_id = $2)
	`
	_, err := s.db.Exec(query, accountId.String(), followerId.String())
	return err
}

func (s *postgresStorage) GetAccountFollowers(accountId uuid.UUID) ([]*Account, error) {
	query := `
		SELECT
			id, username, password, name,
			email, last_login, created_at,
			avatar
		FROM accounts JOIN followers ON accounts.id = followers.follower_id
		WHERE (followers.account_id = $1);
	`
	result, err := s.db.Query(query, accountId.String())
	if err != nil {
		return nil, err
	}

	accounts := []*Account{}

	for result.Next() {
		account := &Account{}
		err := result.Scan(
			&account.ID,
			&account.Username,
			&account.Password,
			&account.Name,
			&account.Email,
			&account.LastLogin,
			&account.CreatedAt,
			&account.Avatar,
		)

		if err != nil {
			return nil, err
		}

		account.Deleted = false
		accounts = append(accounts, account)
	}
	return accounts, nil
}
