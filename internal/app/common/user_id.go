// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
)

const systemUserUniqueUserId = "S0000000"
const unknownUser = int64(0)

func GetCurrentUserId(db database.Database) (int64, error) {
	return getUserIdFromDb(db, systemUserUniqueUserId)
}

func getUserIdFromDb(db database.Database, uniqueUserId string) (int64, error) {
	filter := &entity.UserFilter{UniqueUserID: []*string{&uniqueUserId}}
	ids, err := db.GetAllUserIds(filter)
	if err != nil {
		return unknownUser, fmt.Errorf("Unable to get user ids %w", err)
	} else if len(ids) < 1 {
		return unknownUser, nil
	}
	return ids[0], nil
}
