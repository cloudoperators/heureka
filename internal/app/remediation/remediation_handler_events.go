// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package remediation

import (
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/entity"
)

const (
	ListRemediationsEventName  event.EventName = "ListRemediations"
	CreateRemediationEventName event.EventName = "CreateRemediation"
	UpdateRemediationEventName event.EventName = "UpdateRemediation"
	DeleteRemediationEventName event.EventName = "DeleteRemediation"
)

type ListRemediationsEvent struct {
	Filter       *entity.RemediationFilter
	Options      *entity.ListOptions
	Remediations *entity.List[entity.RemediationResult]
}

func (e *ListRemediationsEvent) Name() event.EventName {
	return ListRemediationsEventName
}

type CreateRemediationEvent struct {
	Remediation *entity.Remediation
}

func (e *CreateRemediationEvent) Name() event.EventName {
	return CreateRemediationEventName
}

type UpdateRemediationEvent struct {
	Remediation *entity.Remediation
}

func (e *UpdateRemediationEvent) Name() event.EventName {
	return UpdateRemediationEventName
}

type DeleteRemediationEvent struct {
	RemediationID int64
}

func (e *DeleteRemediationEvent) Name() event.EventName {
	return DeleteRemediationEventName
}
