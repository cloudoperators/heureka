package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const defaultExpiresInHours = 30 * 24

type TokenClaims struct {
	Version string `json:"version"`
	jwt.RegisteredClaims
}

func GenerateJWT(jwtSecret []byte, expireIn time.Duration) (string, error) {
	claims := TokenClaims{
		Version: "0.3.1",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireIn)),
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
	var jwtSecret []byte
	jwtSecretVar := os.Getenv("JWT_SECRET")
	if jwtSecretVar == "" {
		panic("JWT_SECRET environment variable not set")
	} else {
		jwtSecret = []byte(jwtSecretVar)
	}

	var expiration time.Duration
	expiresInHours, err := strconv.Atoi(os.Getenv("EXPIRES_IN_HOURS"))
	if err != nil {
		expiration = time.Hour * defaultExpiresInHours
	} else {
		expiration = time.Hour * time.Duration(expiresInHours)
	}

	token, err := GenerateJWT(jwtSecret, expiration)
	if err != nil {
		panic(err)
	}

	fmt.Println(token)
}
