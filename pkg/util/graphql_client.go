// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"strings"
	"time"
)

func RequestWithBackoff(requestFunction func() error) error {
	if err := requestFunction(); err != nil {
		if strings.Contains(err.Error(), "connect: connection refused") {
			for i := 0; i < 5; i++ {
				time.Sleep(1 * time.Second)
				if err := requestFunction(); err != nil {
					return err
				}
			}
		} else {
			return err
		}
	}
	return nil
}
