/*
 * Copyright 2020-2024 Luke Whritenour

 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 *     http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package database

import (
	"context"
	"database/sql"
	"net/url"
	"sync"

	"github.com/lukewhrit/spacebin/internal/util"
	_ "modernc.org/sqlite"
)

type SQLite struct {
	*sql.DB
	sync.RWMutex
}

func NewSQLite(uri *url.URL) (Database, error) {
	db, err := sql.Open("sqlite", uri.Host)

	return &SQLite{db, sync.RWMutex{}}, err
}

func (s *SQLite) Migrate(ctx context.Context) error {
	_, err := s.Exec(`
CREATE TABLE IF NOT EXISTS documents (
    id TEXT PRIMARY KEY,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS accounts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL,
	password TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
	public TEXT PRIMARY KEY,
	token TEXT NOT NULL,
	secret TEXT NOT NULL
);`)

	return err
}

func (s *SQLite) GetDocument(ctx context.Context, id string) (Document, error) {
	s.RLock()
	defer s.RUnlock()

	doc := new(Document)
	row := s.QueryRow("SELECT * FROM documents WHERE id=$1", id)
	err := row.Scan(&doc.ID, &doc.Content, &doc.CreatedAt, &doc.UpdatedAt)

	return *doc, err
}

func (s *SQLite) CreateDocument(ctx context.Context, id, content string) error {
	s.Lock()
	defer s.Unlock()

	tx, err := s.Begin()

	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO documents (id, content) VALUES ($1, $2)",
		id, content) // created_at and updated_at are auto-generated

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *SQLite) GetAccount(ctx context.Context, id string) (Account, error) {
	s.RLock()
	defer s.RUnlock()

	acc := new(Account)
	row := s.QueryRow("SELECT * FROM accounts WHERE id=$1", id)
	err := row.Scan(&acc.ID, &acc.Username, &acc.Password)

	return *acc, err
}

func (s *SQLite) GetAccountByUsername(ctx context.Context, username string) (Account, error) {
	account := new(Account)
	row := s.QueryRow("SELECT * FROM accounts WHERE username=$1", username)
	err := row.Scan(&account.ID, &account.Username, &account.Password)

	return *account, err
}

func (s *SQLite) CreateAccount(ctx context.Context, username, password string) error {
	s.Lock()
	defer s.Unlock()

	tx, err := s.Begin()

	if err != nil {
		return err
	}

	// Add account to database
	// Hash and salt the password
	_, err = tx.Exec("INSERT INTO accounts (username, password) VALUES ($1, $2)",
		username, util.HashAndSalt([]byte(password)))

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *SQLite) DeleteAccount(ctx context.Context, id string) error {
	s.Lock()
	defer s.Unlock()

	tx, err := s.Begin()

	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM accounts WHERE id=$1", id)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *SQLite) GetSession(ctx context.Context, id string) (Session, error) {
	s.RLock()
	defer s.RUnlock()

	session := new(Session)
	row := s.QueryRow("SELECT * FROM sessions WHERE id=?", id)
	err := row.Scan(&session.Public, &session.Token, &session.Secret)

	return *session, err
}

func (s *SQLite) CreateSession(ctx context.Context, public, token, secret string) error {
	s.Lock()
	defer s.Unlock()

	tx, err := s.Begin()

	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO sessions (public, token, secret) VALUES ($1, $2, $3)",
		public, token, secret)

	if err != nil {
		return err
	}

	return tx.Commit()
}
