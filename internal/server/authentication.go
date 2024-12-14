package server

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/lukewhrit/spacebin/internal/util"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/sha3"
)

func (s *Server) SignUp(w http.ResponseWriter, r *http.Request) {
	body := util.SignupRequest{
		Username: "luke",
		Password: "password",
	}

	// Do validation
	// Make sure password is secure, make sure username does not exist

	// Create account
	// Encryption handled in Database function
	err := s.Database.CreateAccount(r.Context(), body.Username, body.Password)

	if err != nil {
		util.WriteError(w, http.StatusInternalServerError, err)
	}

	// Respond on success with account ID and username
	account, err := s.Database.GetAccountByUsername(r.Context(), body.Username)

	if err != nil {
		util.WriteError(w, http.StatusInternalServerError, err)
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":       account.ID,
		"username": account.Username,
	})

}

func (s *Server) SignIn(w http.ResponseWriter, r *http.Request) {
	body := &util.SigninRequest{
		Username: "luke",
		Password: "password",
	}

	// if err != nil {
	// 	util.WriteError(w, http.StatusBadRequest, err)
	// }

	// Get user from database
	acc, err := s.Database.GetAccountByUsername(r.Context(), body.Username)

	if err != nil {
		util.WriteError(w, http.StatusInternalServerError, err)
	}

	// Compare passwords
	if bcrypt.CompareHashAndPassword([]byte(acc.Password), []byte(body.Password)) == nil {
		// Generate public, secret keys and salt
		pub, sec, salt, err := util.GenerateStrings([]int{64, 64, 32})

		if err != nil {
			log.Fatal(err)
		}

		// Salt secret key
		buf := []byte(sec + salt)
		secret := make([]byte, 64)
		sha3.ShakeSum256(secret, buf)

		// Create user and server tokens for later comparison
		userToken := util.MakeToken(util.Token{
			Version: "v1",
			Public:  pub,
			Secret:  base64.URLEncoding.EncodeToString([]byte(sec)),
			Salt:    salt,
		})

		serverToken := util.MakeToken(util.Token{
			Version: "v1",
			Public:  pub,
			Secret:  fmt.Sprintf("%x", secret),
			Salt:    salt,
		})

		if err != nil {
			util.WriteError(w, http.StatusInternalServerError, err)
		}

		// Add session to Postgres
		if err := s.Database.CreateSession(r.Context(), pub, userToken, serverToken); err != nil {
			util.WriteError(w, http.StatusInternalServerError, err)
		}

		util.WriteJSON(w, http.StatusOK, map[string]string{
			"token": userToken,
		})
	} else {
		util.WriteError(w, http.StatusUnauthorized, errors.New("invalid username or password"))
	}
}
