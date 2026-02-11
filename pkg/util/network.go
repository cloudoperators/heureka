// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"math/rand"
	"net"
)

func IsPortFree(port string) bool {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return false
	}
	_ = ln.Close()

	return true
}

func GetRandomFreePort() string {
	//nolint:gosec
	rndNumber := rand.Intn(9999)
	port := fmt.Sprintf("2%04d", rndNumber)
	if IsPortFree(port) {
		return port
	}
	return GetRandomFreePort()
}
