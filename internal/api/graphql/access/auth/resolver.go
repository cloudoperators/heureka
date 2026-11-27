// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

// UserResolver looks up a user by their unique_user_id (the JWT/OIDC subject).
// Returns the row's numeric id and type; an error if no row matches.
type UserResolver func(ctx context.Context, uniqueUserId string) (userId int64, userType entity.UserType, err error)
