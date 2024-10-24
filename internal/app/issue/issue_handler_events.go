// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue

import (
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	CreateIssueEventName                     event.EventName = "CreateIssue"
	UpdateIssueEventName                     event.EventName = "UpdateIssue"
	DeleteIssueEventName                     event.EventName = "DeleteIssue"
	AddComponentVersionToIssueEventName      event.EventName = "AddComponentVersionToIssue"
	RemoveComponentVersionFromIssueEventName event.EventName = "RemoveComponentVersionFromIssue"
	ListIssuesEventName                      event.EventName = "ListIssues"
	GetIssueEventName                        event.EventName = "GetIssue"
	ListIssueNamesEventName                  event.EventName = "ListIssueNames"
)

type CreateIssueEvent struct {
	Issue *entity.Issue
}

func (e *CreateIssueEvent) Name() event.EventName {
	return CreateIssueEventName
}

type UpdateIssueEvent struct {
	Issue *entity.Issue
}

func (e *UpdateIssueEvent) Name() event.EventName {
	return UpdateIssueEventName
}

type DeleteIssueEvent struct {
	IssueID int64
}

func (e *DeleteIssueEvent) Name() event.EventName {
	return DeleteIssueEventName
}

type AddComponentVersionToIssueEvent struct {
	IssueID            int64
	ComponentVersionID int64
}

func (e *AddComponentVersionToIssueEvent) Name() event.EventName {
	return AddComponentVersionToIssueEventName
}

type RemoveComponentVersionFromIssueEvent struct {
	IssueID            int64
	ComponentVersionID int64
}

func (e *RemoveComponentVersionFromIssueEvent) Name() event.EventName {
	return RemoveComponentVersionFromIssueEventName
}

type ListIssuesEvent struct {
	Filter  *entity.IssueFilter
	Options *entity.IssueListOptions
	Issues  *entity.IssueList
}

func (e *ListIssuesEvent) Name() event.EventName {
	return ListIssuesEventName
}

type GetIssueEvent struct {
	IssueID int64
	Issue   *entity.Issue
}

func (e *GetIssueEvent) Name() event.EventName {
	return GetIssueEventName
}

type ListIssueNamesEvent struct {
	Filter  *entity.IssueFilter
	Options *entity.ListOptions
	Names   []string
}

func (e *ListIssueNamesEvent) Name() event.EventName {
	return ListIssueNamesEventName
}

// OnComponentVersionAttachmentToIssue is an event handler whenever a ComponentVersion
// is attached to an Issue.
func OnComponentVersionAttachmentToIssue(db database.Database, e event.Event) {
	l := logrus.WithFields(logrus.Fields{
		"event":   "OnComponentVersionAttachmentToIssue",
		"payload": e,
	})

	if attachmentEvent, ok := e.(*AddComponentVersionToIssueEvent); ok {
		l.WithField("event-step", "GetComponentInstances").Debug("Get Component Instances by ComponentVersionId")
		componentInstances, err := db.GetComponentInstances(&entity.ComponentInstanceFilter{
			ComponentVersionId: []*int64{&attachmentEvent.ComponentVersionID},
		})

		if err != nil {
			l.WithField("event-step", "GetComponentInstances").WithError(err).Error("Error while fetching ComponentInstances")
			return
		}

		for _, compInst := range componentInstances {
			l.WithField("event-step", "GetIssueMatches").Debug("Fetching issue matches related to Component Instance")
			issue_matches, err := db.GetIssueMatches(&entity.IssueMatchFilter{
				ComponentInstanceId: []*int64{&compInst.Id},
			})

			if err != nil {
				l.WithField("event-step", "GetIssueMatches").WithError(err).Error("Error while fetching issue matches related to Component Instance")
				return
			}
			l.WithField("issueMatchesCount", len(issue_matches))

			// If the issue match is already created, we ignore it, as we do not associate with a version change a severity change
			if len(issue_matches) != 0 {
				l.WithField("event-step", "Skipping").Debug("The issue match does already exist. Skipping")
				continue
			}

			// Create new issue match
			// TODO: Implement this properly
			issue_match := &entity.IssueMatch{
				UserId:                1, // TODO: change this?
				Status:                entity.IssueMatchStatusValuesNew,
				Severity:              issueVariantMap[issueId].Severity, //we got two  simply take the first one
				ComponentInstanceId:   compInst.Id,
				IssueId:               attachmentEvent.IssueID,
				TargetRemediationDate: GetTargetRemediationTimeline(issueVariant.Severity, time.Now()),
			}
			l.WithField("event-step", "CreateIssueMatch").WithField("issueMatch", issue_match).Debug("Creating Issue Match")

			_, err = db.CreateIssueMatch(issue_match)
			if err != nil {
				l.WithField("event-step", "CreateIssueMatch").WithError(err).Error("Error while creating issue match")
				return
			}

		}
	} else {
		l.Error("Wrong event")
	}

}
