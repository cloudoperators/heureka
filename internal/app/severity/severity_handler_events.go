// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package severity

import (
	"github.wdf.sap.corp/cc/heureka/internal/app/event"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

const (
	GetSeverityEventName event.EventName = "GetSeverity"
)

type GetSeverityEvent struct {
	Filter *entity.SeverityFilter
	Result *entity.Severity
}

func (e *GetSeverityEvent) Name() event.EventName {
	return GetSeverityEventName
}