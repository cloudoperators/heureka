package issue_repository

import (
	"github.wdf.sap.corp/cc/heureka/internal/app/event"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

const (
	ListIssueRepositoriesEventName event.EventName = "ListIssueRepositories"
	CreateIssueRepositoryEventName event.EventName = "CreateIssueRepository"
	UpdateIssueRepositoryEventName event.EventName = "UpdateIssueRepository"
	DeleteIssueRepositoryEventName event.EventName = "DeleteIssueRepository"
)

type ListIssueRepositoriesEvent struct {
	Filter  *entity.IssueRepositoryFilter
	Options *entity.ListOptions
	Results *entity.List[entity.IssueRepositoryResult]
}

func (e *ListIssueRepositoriesEvent) Name() event.EventName {
	return ListIssueRepositoriesEventName
}

type CreateIssueRepositoryEvent struct {
	IssueRepository *entity.IssueRepository
}

func (e *CreateIssueRepositoryEvent) Name() event.EventName {
	return CreateIssueRepositoryEventName
}

type UpdateIssueRepositoryEvent struct {
	IssueRepository *entity.IssueRepository
}

func (e *UpdateIssueRepositoryEvent) Name() event.EventName {
	return UpdateIssueRepositoryEventName
}

type DeleteIssueRepositoryEvent struct {
	IssueRepositoryID int64
}

func (e *DeleteIssueRepositoryEvent) Name() event.EventName {
	return DeleteIssueRepositoryEventName
}
