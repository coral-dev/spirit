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

package util

import "errors"

var ErrTooManyParts = errors.New("ratelimiter string invalid: too many parts")

// DocumentResponse is a document object
type DocumentResponse struct {
	ID        string `json:"id,omitempty"`         // The document ID.
	Content   string `json:"content,omitempty"`    // The document content.
	CreatedAt int64  `json:"created_at,omitempty"` // The Unix timestamp of when the document was inserted.
	UpdatedAt int64  `json:"updated_at,omitempty"` // The Unix timestamp of when the document was last modified.
	Exists    bool   `json:"exists,omitempty"`     // Whether the document does or does not exist.
}

// Token is an authentication token object
type Token struct {
	Version string
	Public  string
	Secret  string
	Salt    string
}

// CreateRequest represents a POST request to create a document
type CreateRequest struct {
	Content string
}

// SigninRequest represents a POST request to authenticate an account
type SigninRequest struct {
	Username string
	Password string
}

// SignupRequest represents a POST request to register an account
type SignupRequest struct {
	Username string
	Password string
}
