// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package evidence

import (
	"encoding/json"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/event"
)

const (
	ListEvidencesEventName  event.EventName = "ListEvidences"
	CreateEvidenceEventName event.EventName = "CreateEvidence"
	UpdateEvidenceEventName event.EventName = "UpdateEvidence"
	DeleteEvidenceEventName event.EventName = "DeleteEvidence"
)

type ListEvidencesEvent struct {
	Filter  *entity.EvidenceFilter
	Options *entity.ListOptions
	Results *entity.List[entity.EvidenceResult]
}

func (e ListEvidencesEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &ListEvidencesEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *ListEvidencesEvent) Name() event.EventName {
	return ListEvidencesEventName
}

type CreateEvidenceEvent struct {
	Evidence *entity.Evidence
}

func (e CreateEvidenceEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &CreateEvidenceEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *CreateEvidenceEvent) Name() event.EventName {
	return CreateEvidenceEventName
}

type UpdateEvidenceEvent struct {
	Evidence *entity.Evidence
}

func (e UpdateEvidenceEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &UpdateEvidenceEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *UpdateEvidenceEvent) Name() event.EventName {
	return UpdateEvidenceEventName
}

type DeleteEvidenceEvent struct {
	EvidenceID int64
}

func (e DeleteEvidenceEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &DeleteEvidenceEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *DeleteEvidenceEvent) Name() event.EventName {
	return DeleteEvidenceEventName
}
