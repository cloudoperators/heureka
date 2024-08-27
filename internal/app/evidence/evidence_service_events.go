package evidence

import (
	"github.wdf.sap.corp/cc/heureka/internal/app/event"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
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

func (e *ListEvidencesEvent) Name() event.EventName {
	return ListEvidencesEventName
}

type CreateEvidenceEvent struct {
	Evidence *entity.Evidence
}

func (e *CreateEvidenceEvent) Name() event.EventName {
	return CreateEvidenceEventName
}

type UpdateEvidenceEvent struct {
	Evidence *entity.Evidence
}

func (e *UpdateEvidenceEvent) Name() event.EventName {
	return UpdateEvidenceEventName
}

type DeleteEvidenceEvent struct {
	EvidenceID int64
}

func (e *DeleteEvidenceEvent) Name() event.EventName {
	return DeleteEvidenceEventName
}
