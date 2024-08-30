package main

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.wdf.sap.corp/cc/heureka/internal/api/graphql/access"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

type TokenClaims struct {
	Name    string `json:"name"`
	Role    string `json:"role"`
	Version string `json:"version"`
	jwt.RegisteredClaims
}

func GenerateJWT() (string, error) {
	claims := access.TokenClaims{Name: "a", Role: "b", Version: "c"}
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
