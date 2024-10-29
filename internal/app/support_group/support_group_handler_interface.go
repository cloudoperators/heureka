// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package support_group

import "github.com/cloudoperators/heureka/internal/entity"

type SupportGroupHandler interface {
	ListSupportGroups(*entity.SupportGroupFilter, *entity.ListOptions) (*entity.List[entity.SupportGroupResult], error)
	GetSupportGroup(int64) (*entity.SupportGroup, error)
	CreateSupportGroup(*entity.SupportGroup) (*entity.SupportGroup, error)
	UpdateSupportGroup(*entity.SupportGroup) (*entity.SupportGroup, error)
	DeleteSupportGroup(int64) error
	AddServiceToSupportGroup(int64, int64) (*entity.SupportGroup, error)
	RemoveServiceFromSupportGroup(int64, int64) (*entity.SupportGroup, error)
	AddUserToSupportGroup(int64, int64) (*entity.SupportGroup, error)
	RemoveUserFromSupportGroup(int64, int64) (*entity.SupportGroup, error)
	ListSupportGroupCcrns(*entity.SupportGroupFilter, *entity.ListOptions) ([]string, error)
}
