// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
)

func GetCurrentUserId(db database.Database) (int64, error) {
	return getUserIdFromDb(db, "S0000000")
}

func getUserIdFromDb(db database.Database, uniqueUserId string) (int64, error) {
	filter := &entity.UserFilter{UniqueUserID: []*string{&uniqueUserId}}
	ids, err := db.GetAllUserIds(filter)
	if err != nil {
		return 0, fmt.Errorf("Unable to get user ids %w", err)
	} else if len(ids) < 1 {
		return 0, nil
	}
	return ids[0], nil
}
