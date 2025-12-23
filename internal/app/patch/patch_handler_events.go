// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package patch

import (
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/entity"
)

const (
	ListPatchesEventName event.EventName = "ListPatches"
)

type ListPatchesEvent struct {
	Filter  *entity.PatchFilter
	Options *entity.ListOptions
	Patches *entity.List[entity.PatchResult]
}

func (e *ListPatchesEvent) Name() event.EventName {
	return ListPatchesEventName
}
