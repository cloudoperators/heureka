package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

type TokenClaims struct {
	Version string `json:"version"`
	jwt.RegisteredClaims
}

func GenerateJWT() (string, error) {
	claims := TokenClaims{
		Version: "0.3.1",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "heureka",
			Subject:   "testUser",
        },
    }
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func main() {
	if os.Getenv("JWT_SECRET") == "" {
		panic("JWT_SECRET environment variable not set")
	}

	token, err := GenerateJWT()
	if err != nil {
		panic(err)
	}

	fmt.Println(token)
}
