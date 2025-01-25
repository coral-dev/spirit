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
	"time"

	_ "github.com/lib/pq"
)

type Document struct {
	ID        string    `db:"id" json:"id"`
	Content   string    `db:"content" json:"content"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type Account struct {
	ID       int    `db:"id" json:"id"`
	Username string `db:"username" json:"username"`
	Password string `db:"password" json:"password"`
	// Documents []Document `db:"documents" json:"documents"`
}

type Session struct {
	Public string `db:"public" json:"public"`
	Token  string `db:"token" json:"token"`
	Secret string `db:"secret" json:"secret"`
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Database
type Database interface {
	Migrate(ctx context.Context) error
	Close() error

	GetDocument(ctx context.Context, id string) (Document, error)
	CreateDocument(ctx context.Context, id, content string) error

	GetAccount(ctx context.Context, id string) (Account, error)
	GetAccountByUsername(ctx context.Context, username string) (Account, error)
	CreateAccount(ctx context.Context, username, password string) error
	// UpdateAccount(ctx context.Context, id, username, password string) error
	DeleteAccount(ctx context.Context, id string) error

	GetSession(ctx context.Context, id string) (Session, error)
	CreateSession(ctx context.Context, public, token, secret string) error
}
