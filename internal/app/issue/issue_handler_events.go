// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue

import (
	"github.com/cloudoperators/heureka/internal/event"
	"time"

	"encoding/json"
	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/shared"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

const (
	CreateIssueEventName                     event.EventName = "CreateIssue"
	UpdateIssueEventName                     event.EventName = "UpdateIssue"
	DeleteIssueEventName                     event.EventName = "DeleteIssue"
	AddComponentVersionToIssueEventName      event.EventName = "AddComponentVersionToIssue"
	RemoveComponentVersionFromIssueEventName event.EventName = "RemoveComponentVersionFromIssue"
	ListIssuesEventName                      event.EventName = "ListIssues"
	GetIssueEventName                        event.EventName = "GetIssue"
	GetIssueSeverityCountsEventName          event.EventName = "GetIssueSeverityCounts"
	ListIssueNamesEventName                  event.EventName = "ListIssueNames"
)

type CreateIssueEvent struct {
	Issue *entity.Issue
}

func (e CreateIssueEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &CreateIssueEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *CreateIssueEvent) Name() event.EventName {
	return CreateIssueEventName
}

type UpdateIssueEvent struct {
	Issue *entity.Issue
}

func (e UpdateIssueEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &UpdateIssueEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *UpdateIssueEvent) Name() event.EventName {
	return UpdateIssueEventName
}

type DeleteIssueEvent struct {
	IssueID int64
}

func (e DeleteIssueEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &DeleteIssueEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *DeleteIssueEvent) Name() event.EventName {
	return DeleteIssueEventName
}

type AddComponentVersionToIssueEvent struct {
	IssueID            int64
	ComponentVersionID int64
}

func (e AddComponentVersionToIssueEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &AddComponentVersionToIssueEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *AddComponentVersionToIssueEvent) Name() event.EventName {
	return AddComponentVersionToIssueEventName
}

type RemoveComponentVersionFromIssueEvent struct {
	IssueID            int64
	ComponentVersionID int64
}

func (e RemoveComponentVersionFromIssueEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &RemoveComponentVersionFromIssueEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *RemoveComponentVersionFromIssueEvent) Name() event.EventName {
	return RemoveComponentVersionFromIssueEventName
}

type ListIssuesEvent struct {
	Filter  *entity.IssueFilter
	Options *entity.IssueListOptions
	Issues  *entity.IssueList
}

func (e ListIssuesEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &ListIssuesEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *ListIssuesEvent) Name() event.EventName {
	return ListIssuesEventName
}

type GetIssueEvent struct {
	IssueID int64
	Issue   *entity.Issue
}

func (e GetIssueEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &GetIssueEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *GetIssueEvent) Name() event.EventName {
	return GetIssueEventName
}

type ListIssueNamesEvent struct {
	Filter  *entity.IssueFilter
	Options *entity.ListOptions
	Names   []string
}

func (e ListIssueNamesEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &ListIssueNamesEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *ListIssueNamesEvent) Name() event.EventName {
	return ListIssueNamesEventName
}

type GetIssueSeverityCountsEvent struct {
	Filter *entity.IssueFilter
	Counts *entity.IssueSeverityCounts
}

func (e GetIssueSeverityCountsEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &GetIssueSeverityCountsEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *GetIssueSeverityCountsEvent) Name() event.EventName {
	return GetIssueSeverityCountsEventName
}

// OnComponentVersionAttachmentToIssue is an event handler whenever a ComponentVersion
// is attached to an Issue.
func OnComponentVersionAttachmentToIssue(db database.Database, e event.Event) {
	l := logrus.WithFields(logrus.Fields{
		"event":   "OnComponentVersionAttachmentToIssue",
		"payload": e,
	})

	if attachmentEvent, ok := e.(*AddComponentVersionToIssueEvent); ok {
		// Get ComponentInstances
		l.WithField("event-step", "GetComponentInstances").Debug("Get Component Instances by ComponentVersionId")
		componentInstances, err := db.GetComponentInstances(&entity.ComponentInstanceFilter{
			ComponentVersionId: []*int64{&attachmentEvent.ComponentVersionID},
		})

		if err != nil {
			l.WithField("event-step", "GetComponentInstances").WithError(err).Error("Error while fetching ComponentInstances")
			return
		}

		// For each ComponentInstance get available IssueVariants
		// via GetServiceIssueVariants
		for _, compInst := range componentInstances {
			// Get Service Issue Variants
			issueVariantMap, err := shared.BuildIssueVariantMap(db, &entity.ServiceIssueVariantFilter{
				ComponentInstanceId: []*int64{&compInst.Id},
				IssueId:             []*int64{&attachmentEvent.IssueID},
			}, attachmentEvent.ComponentVersionID)
			if err != nil {
				l.WithField("event-step", "FetchIssueVariants").WithError(err).Error("Error while fetching issue variants")
			}

			// Create new IssueMatches
			createIssueMatches(db, l, compInst.Id, issueVariantMap)
		}
	} else {
		l.Error("Invalid event type received")
	}

}

// TODO: This function is very similar to the one used in issue_match_handler_events.go
// We might as well put this into the shared package
//
// createIssueMatches creates new issue matches based on the component instance Id,
// issue ID and their corresponding issue variants (sorted by priority)
func createIssueMatches(
	db database.Database,
	l *logrus.Entry,
	componentInstanceId int64,
	issueVariantMap map[int64]entity.ServiceIssueVariant,
) {
	for issueId, issueVariant := range issueVariantMap {
		l = l.WithFields(logrus.Fields{
			"issue": issueVariant,
		})

		// Check if IssueMatches already exist
		l.WithField("event-step", "GetIssueMatches").Debug("Fetching issue matches related to assigned Component Instance")
		issue_matches, err := db.GetIssueMatches(&entity.IssueMatchFilter{
			IssueId:             []*int64{&issueId},
			ComponentInstanceId: []*int64{&componentInstanceId},
		}, []entity.Order{})

		if err != nil {
			l.WithField("event-step", "FetchIssueMatches").WithError(err).Error("Error while fetching issue matches related to assigned Component Instance")
		}
		l.WithField("issueMatchesCount", len(issue_matches))

		if len(issue_matches) != 0 {
			l.WithField("event-step", "Skipping").Debug("The issue match does already exist. Skipping")
			continue
		}

		user, err := common.GetCurrentUserId(db)
		if err != nil {
			l.WithField("event-step", "GetCurrentUserId").WithError(err).Error("Error while getting current user ID")
			continue
		}

		// Create new issue match
		issue_match := &entity.IssueMatch{
			Metadata: entity.Metadata{
				CreatedBy: user,
				UpdatedBy: user,
			},
			UserId:                user,
			Status:                entity.IssueMatchStatusValuesNew,
			Severity:              issueVariantMap[issueId].Severity, //we got two  simply take the first one
			ComponentInstanceId:   componentInstanceId,
			IssueId:               issueId,
			TargetRemediationDate: shared.GetTargetRemediationTimeline(issueVariant.Severity, time.Now(), nil),
		}
		l.WithField("event-step", "CreateIssueMatch").WithField("issueMatch", issue_match).Debug("Creating Issue Match")

		_, err = db.CreateIssueMatch(issue_match)
		if err != nil {
			l.WithField("event-step", "CreateIssueMatch").WithError(err).Error("Error while creating issue match")
			continue
		}
	}
}
