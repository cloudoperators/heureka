// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/cloudoperators/heureka/pkg/jwt"
)

const defaultExpiresInHours = 30 * 24

func main() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET environment variable not set")
	}

	expiration := time.Hour * defaultExpiresInHours
	if hours, err := strconv.Atoi(os.Getenv("EXPIRES_IN_HOURS")); err == nil {
		expiration = time.Hour * time.Duration(hours)
	}

	token, err := jwt.GenerateJWT([]byte(secret), expiration, "testclient")
	if err != nil {
		panic(err)
	}
	fmt.Println(token)
}
