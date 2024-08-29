package main

import (
	"fmt"
	"os"

	"github.com/dgrijalva/jwt-go"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func GenerateJWT() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"scanner_name": "ttestscanner",
		//"scanner_name": "testscanner",
		"jwt_version":  "0.1.0",
	})

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
