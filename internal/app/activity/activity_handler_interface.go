// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package activity

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

type ActivityHandler interface {
	ListActivities(*entity.ActivityFilter, *entity.ListOptions) (*entity.List[entity.ActivityResult], error)
	GetActivity(int64) (*entity.Activity, error)
	CreateActivity(context.Context, *entity.Activity) (*entity.Activity, error)
	UpdateActivity(context.Context, *entity.Activity) (*entity.Activity, error)
	DeleteActivity(context.Context, int64) error
	AddServiceToActivity(int64, int64) (*entity.Activity, error)
	RemoveServiceFromActivity(int64, int64) (*entity.Activity, error)
	AddIssueToActivity(int64, int64) (*entity.Activity, error)
	RemoveIssueFromActivity(int64, int64) (*entity.Activity, error)
}
