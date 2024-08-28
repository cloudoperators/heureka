// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package activity

import "github.wdf.sap.corp/cc/heureka/internal/entity"

type ActivityHandler interface {
	ListActivities(*entity.ActivityFilter, *entity.ListOptions) (*entity.List[entity.ActivityResult], error)
	GetActivity(int64) (*entity.Activity, error)
	CreateActivity(*entity.Activity) (*entity.Activity, error)
	UpdateActivity(*entity.Activity) (*entity.Activity, error)
	DeleteActivity(int64) error
	AddServiceToActivity(int64, int64) (*entity.Activity, error)
	RemoveServiceFromActivity(int64, int64) (*entity.Activity, error)
	AddIssueToActivity(int64, int64) (*entity.Activity, error)
	RemoveIssueFromActivity(int64, int64) (*entity.Activity, error)
}
