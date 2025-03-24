// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package severity

import (
	"encoding/json"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/event"
)

const (
	GetSeverityEventName event.EventName = "GetSeverity"
)

type GetSeverityEvent struct {
	Filter *entity.SeverityFilter
	Result *entity.Severity
}

func (e GetSeverityEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &GetSeverityEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *GetSeverityEvent) Name() event.EventName {
	return GetSeverityEventName
}
