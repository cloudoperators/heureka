// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package user

import "github.com/cloudoperators/heureka/internal/entity"

type UserHandler interface {
	ListUsers(*entity.UserFilter, *entity.ListOptions) (*entity.List[entity.UserResult], error)
	CreateUser(*entity.User) (*entity.User, error)
	UpdateUser(*entity.User) (*entity.User, error)
	DeleteUser(int64) error
	ListUserNames(*entity.UserFilter, *entity.ListOptions) ([]string, error)
	ListUniqueUserIDs(*entity.UserFilter, *entity.ListOptions) ([]string, error)
}
