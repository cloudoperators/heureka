// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package support_group

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

type SupportGroupHandler interface {
	ListSupportGroups(
		context.Context,
		*entity.SupportGroupFilter,
		*entity.ListOptions,
	) (*entity.List[entity.SupportGroupResult], error)
	GetSupportGroup(context.Context, int64) (*entity.SupportGroup, error)
	CreateSupportGroup(context.Context, *entity.SupportGroup) (*entity.SupportGroup, error)
	UpdateSupportGroup(context.Context, *entity.SupportGroup) (*entity.SupportGroup, error)
	DeleteSupportGroup(context.Context, int64) error
	AddServiceToSupportGroup(context.Context, int64, int64) (*entity.SupportGroup, error)
	RemoveServiceFromSupportGroup(context.Context, int64, int64) (*entity.SupportGroup, error)
	AddUserToSupportGroup(context.Context, int64, int64) (*entity.SupportGroup, error)
	RemoveUserFromSupportGroup(context.Context, int64, int64) (*entity.SupportGroup, error)
	ListSupportGroupCcrns(*entity.SupportGroupFilter, *entity.ListOptions) ([]string, error)
}
