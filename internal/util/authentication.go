package util

import (
	"crypto/rand"
	"fmt"
	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func HashAndSalt(pwd []byte) string {
	// Use GenerateFromPassword to hash & salt pwd.
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Fatalln(err)
	} // GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash)
}

func PrngString() (string, error) {
	b := make([]byte, 10)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", b), nil
}

func GenerateStrings(bits []int) (a, b, c string, err error) {
	if a, err = PrngString(); err != nil {
		return "", "", "", err
	}

	if b, err = PrngString(); err != nil {
		return "", "", "", err
	}

	if c, err = PrngString(); err != nil {
		return "", "", "", err
	}

	return a, b, c, err
}

func ParseToken(token string) (Token, error) {
	var tok Token
	toks := strings.Split(token, ".")

	tok.Version = toks[0]
	tok.Public = toks[1]
	tok.Secret = toks[2]

	if len(toks) == 4 {
		tok.Salt = toks[3]
	}

	return tok, nil
}

func MakeToken(token Token) string {
	return fmt.Sprintf("%s.%s.%s.%s", token.Version, token.Public, token.Secret, token.Salt)
}
