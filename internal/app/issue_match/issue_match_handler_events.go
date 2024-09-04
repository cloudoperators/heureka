// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_match

import (
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/entity"
)

const (
	ListIssueMatchesEventName             event.EventName = "ListIssueMatches"
	GetIssueMatchEventName                event.EventName = "GetIssueMatch"
	CreateIssueMatchEventName             event.EventName = "CreateIssueMatch"
	UpdateIssueMatchEventName             event.EventName = "UpdateIssueMatch"
	DeleteIssueMatchEventName             event.EventName = "DeleteIssueMatch"
	AddEvidenceToIssueMatchEventName      event.EventName = "AddEvidenceToIssueMatch"
	RemoveEvidenceFromIssueMatchEventName event.EventName = "RemoveEvidenceFromIssueMatch"
)

type ListIssueMatchesEvent struct {
	Filter  *entity.IssueMatchFilter
	Options *entity.ListOptions
	Results *entity.List[entity.IssueMatchResult]
}

func (e *ListIssueMatchesEvent) Name() event.EventName {
	return ListIssueMatchesEventName
}

type GetIssueMatchEvent struct {
	IssueMatchID int64
	Result       *entity.IssueMatch
}

func (e *GetIssueMatchEvent) Name() event.EventName {
	return GetIssueMatchEventName
}

type CreateIssueMatchEvent struct {
	IssueMatch *entity.IssueMatch
}

func (e *CreateIssueMatchEvent) Name() event.EventName {
	return CreateIssueMatchEventName
}

type UpdateIssueMatchEvent struct {
	IssueMatch *entity.IssueMatch
}

func (e *UpdateIssueMatchEvent) Name() event.EventName {
	return UpdateIssueMatchEventName
}

type DeleteIssueMatchEvent struct {
	IssueMatchID int64
}

func (e *DeleteIssueMatchEvent) Name() event.EventName {
	return DeleteIssueMatchEventName
}

type AddEvidenceToIssueMatchEvent struct {
	IssueMatchID int64
	EvidenceID   int64
}

func (e *AddEvidenceToIssueMatchEvent) Name() event.EventName {
	return AddEvidenceToIssueMatchEventName
}

type RemoveEvidenceFromIssueMatchEvent struct {
	IssueMatchID int64
	EvidenceID   int64
}

func (e *RemoveEvidenceFromIssueMatchEvent) Name() event.EventName {
	return RemoveEvidenceFromIssueMatchEventName
}
