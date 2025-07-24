// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"strings"
)

type KeyHashType int

const (
	KEY_HASH_BASE64 KeyHashType = iota
	KEY_HASH_SHA256
	KEY_HASH_SHA512
	KEY_HASH_HEX
	KEY_HASH_NONE
)

const DefaultKeyHash = KEY_HASH_BASE64

func (k KeyHashType) String() string {
	switch k {
	case KEY_HASH_BASE64:
		return "Base64"
	case KEY_HASH_SHA256:
		return "SHA256"
	case KEY_HASH_SHA512:
		return "SHA512"
	case KEY_HASH_HEX:
		return "HEX"
	case KEY_HASH_NONE:
		return "None"
	default:
		return "Unknown"
	}
}

func ParseKeyHashType(s string) KeyHashType {
	switch strings.ToLower(s) {
	case "base64":
		return KEY_HASH_BASE64
	case "sha256":
		return KEY_HASH_SHA256
	case "sha512":
		return KEY_HASH_SHA512
	case "hex":
		return KEY_HASH_HEX
	case "none":
		return KEY_HASH_NONE
	default:
		return DefaultKeyHash
	}
}

func encodeBase64(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

func encodeSHA256(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

func encodeSHA512(input string) string {
	hash := sha512.Sum512([]byte(input))
	return hex.EncodeToString(hash[:])
}

func encodeHex(input string) string {
	return hex.EncodeToString([]byte(input))
}

func decodeBase64(encoded string) (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decodedBytes), nil
}

func decodeHex(hexStr string) (string, error) {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
