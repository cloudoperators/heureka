// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package util

import (
	cr "crypto/rand"
	"math/big"
)

func GenerateRandomString(length int, cs *string) string {
	// charset for random string
	var charset string
	if cs == nil {
		charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	} else {
		charset = *cs
	}
	// Create a byte slice to store the random string
	randomBytes := make([]byte, length)
	// Calculate the maximum index in the charset
	maxIndex := big.NewInt(int64(len(charset)))
	// Generate random bytes and map them to characters in the charset
	for i := 0; i < length; i++ {
		randomIndex, _ := cr.Int(cr.Reader, maxIndex)
		randomBytes[i] = charset[randomIndex.Int64()]
	}
	// Convert the byte slice to a string and return it
	return string(randomBytes)
}
