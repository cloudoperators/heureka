package issue_match_change

import (
	"github.wdf.sap.corp/cc/heureka/internal/app/event"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

const (
	ListIssueMatchChangesEventName  event.EventName = "ListIssueMatchChanges"
	CreateIssueMatchChangeEventName event.EventName = "CreateIssueMatchChange"
	UpdateIssueMatchChangeEventName event.EventName = "UpdateIssueMatchChange"
	DeleteIssueMatchChangeEventName event.EventName = "DeleteIssueMatchChange"
)

type ListIssueMatchChangesEvent struct {
	Filter  *entity.IssueMatchChangeFilter
	Options *entity.ListOptions
	Results *entity.List[entity.IssueMatchChangeResult]
}

func (e *ListIssueMatchChangesEvent) Name() event.EventName {
	return ListIssueMatchChangesEventName
}

type CreateIssueMatchChangeEvent struct {
	IssueMatchChange *entity.IssueMatchChange
}

func (e *CreateIssueMatchChangeEvent) Name() event.EventName {
	return CreateIssueMatchChangeEventName
}

type UpdateIssueMatchChangeEvent struct {
	IssueMatchChange *entity.IssueMatchChange
}

func (e *UpdateIssueMatchChangeEvent) Name() event.EventName {
	return UpdateIssueMatchChangeEventName
}

type DeleteIssueMatchChangeEvent struct {
	IssueMatchChangeID int64
}

func (e *DeleteIssueMatchChangeEvent) Name() event.EventName {
	return DeleteIssueMatchChangeEventName
}
