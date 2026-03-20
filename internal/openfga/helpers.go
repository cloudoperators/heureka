// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package openfga

import (
	"strconv"

	"github.com/sirupsen/logrus"
)

// ObjectIdFromInt converts a numeric ID to an OpenFGA ObjectId.
func ObjectIdFromInt(id int64) ObjectId {
	return ObjectId(strconv.FormatInt(id, 10))
}

// UserIdFromInt converts an int ID to an OpenFGA UserId.
func UserIdFromInt(id int64) UserId {
	return UserId(strconv.FormatInt(id, 10))
}

// helper: build a consistent log entry for a relation input
func (a *Authz) logRel(event string, r RelationInput) *logrus.Entry {
	return a.logger.WithFields(logrus.Fields{
		"event":      event,
		"userType":   r.UserType,
		"user":       r.UserId,
		"relation":   r.Relation,
		"objectType": r.ObjectType,
		"objectId":   r.ObjectId,
	})
}
