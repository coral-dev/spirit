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

	_ "github.com/lib/pq"
	"github.com/lukewhrit/spacebin/internal/util"
)

type Postgres struct {
	*sql.DB
}

func NewPostgres(uri *url.URL) (Database, error) {
	db, err := sql.Open("postgres", uri.String())

	return &Postgres{db}, err
}

func (p *Postgres) Migrate(ctx context.Context) error {
	_, err := p.Exec(`
CREATE TABLE IF NOT EXISTS documents (
	id varchar(255) PRIMARY KEY,
	content text NOT NULL,
	created_at timestamp with time zone DEFAULT now(),
	updated_at timestamp with time zone DEFAULT now()
);

CREATE TABLE IF NOT EXISTS accounts (
	id SERIAL PRIMARY KEY,
	username varchar(255) NOT NULL,
	password varchar(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
	public varchar(255) PRIMARY KEY,
	token varchar(255) NOT NULL,
	secret varchar
);`)

	return err
}

func (p *Postgres) GetDocument(ctx context.Context, id string) (Document, error) {
	doc := new(Document)
	row := p.QueryRow("SELECT * FROM documents WHERE id=$1", id)
	err := row.Scan(&doc.ID, &doc.Content, &doc.CreatedAt, &doc.UpdatedAt)

	return *doc, err
}

func (p *Postgres) CreateDocument(ctx context.Context, id, content string) error {
	tx, err := p.Begin()

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

func (p *Postgres) GetAccount(ctx context.Context, id string) (Account, error) {
	account := new(Account)
	row := p.QueryRow("SELECT * FROM accounts WHERE id=$1", id)
	err := row.Scan(&account.ID, &account.Username, &account.Password)

	return *account, err
}

func (p *Postgres) GetAccountByUsername(ctx context.Context, username string) (Account, error) {
	account := new(Account)
	row := p.QueryRow("SELECT * FROM accounts WHERE username=$1", username)
	err := row.Scan(&account.ID, &account.Username, &account.Password)

	return *account, err
}

func (p *Postgres) CreateAccount(ctx context.Context, username, password string) error {
	tx, err := p.Begin()

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

func (p *Postgres) DeleteAccount(ctx context.Context, id string) error {
	tx, err := p.Begin()

	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM accounts WHERE id=$1", id)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (p *Postgres) GetSession(ctx context.Context, id string) (Session, error) {
	session := new(Session)
	row := p.QueryRow("SELECT * FROM sessions WHERE id=$1", id)
	err := row.Scan(&session.Public, &session.Token, &session.Secret)

	return *session, err
}

func (p *Postgres) CreateSession(ctx context.Context, public, token, secret string) error {
	tx, err := p.Begin()

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
