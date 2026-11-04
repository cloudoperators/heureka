// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"
)

func (s *SqlDatabase) Autoclean(ctx context.Context, keepNRuns uint64) error {
	return nil
}
