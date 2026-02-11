// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"
	"fmt"

	authentication_context "github.com/cloudoperators/heureka/internal/api/graphql/access/context"

	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
)

const (
	systemUserUniqueUserId = "S0000000"
	unknownUser            = int64(0)
)

func GetCurrentUserId(ctx context.Context, db database.Database) (int64, error) {
	if authentication_context.IsAuthenticationRequired(ctx) {
		uniqueUserId, err := authentication_context.UserNameFromContext(ctx)
		if err != nil {
			return 0, fmt.Errorf("could not get user name from context: %w", err)
		}

		return getUserIdFromDb(db, uniqueUserId)
	} else {
		return getUserIdFromDb(db, systemUserUniqueUserId)
	}
}

func getUserIdFromDb(db database.Database, uniqueUserId string) (int64, error) {
	filter := &entity.UserFilter{UniqueUserID: []*string{&uniqueUserId}}
	ids, err := db.GetAllUserIds(filter)
	if err != nil {
		return unknownUser, fmt.Errorf("unable to get user ids %w", err)
	} else if len(ids) < 1 {
		return unknownUser, nil
	}
	return ids[0], nil
}

func NewAdminContext() context.Context {
	return context.Background()
}
