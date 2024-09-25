// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_match

import (
	"time"

	"github.com/cloudoperators/heureka/internal/app/component_instance"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
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

func OnComponentInstanceCreate(db database.Database, event event.Event) {
	if createEvent, ok := event.(*component_instance.CreateComponentInstanceEvent); ok {
		OnComponentVersionAssignmentToComponentInstance(db, createEvent.ComponentInstance.Id, createEvent.ComponentInstance.ComponentVersionId)
	}
}

// BuildIssueVariantMap builds a map of issue id to issue variant for the given issues and component instance id
//
//   - It fetches the services and issue repositories related to the component instance
//   - identifies the issue repositories with the highest priority
//   - fetches the issue variants for this repositories and the given issues
//   - and finally creates a map of the relevant issueVariants where in case of multiple issue variants the highest
//     severity variant is taken. In case of multiple variants of the same severity the first one is taken.
//
// @Todo DISCUSS may move getting issues here as well and iterate in the calling function over the issueVariantMap....
// @Todo DISCUSS function still long but mainly due to getting data and logging steps, congnitive complexity is not that high here
//
// @Todo DISCUSS this is essential business logic, does it belong here or should it live somewhere else?
func BuildIssueVariantMap(db database.Database, componentInstanceID int64, componentVersionID int64) (map[int64]entity.IssueVariant, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":               "BuildIssueVariantMap",
		"componentInstanceID": componentInstanceID,
		"componentVersionID":  componentVersionID,
	})
	// Get Issues based on ComponentVersion ID
	issues, err := db.GetIssues(&entity.IssueFilter{ComponentVersionId: []*int64{&componentVersionID}})
	if err != nil {
		l.WithField("event-step", "FetchIssues").WithError(err).Error("Error while fetching issues related to component Version")
		return nil, err
	}

	// Get Services based on ComponentInstance ID
	services, err := db.GetServices(&entity.ServiceFilter{ComponentInstanceId: []*int64{&componentInstanceID}})
	if err != nil {
		l.WithField("event-step", "FetchServices").WithError(err).Error("Error while fetching services related to Component Instance")
		return nil, err
	}

	// Get Issue Repositories
	serviceIds := lo.Map(services, func(service entity.Service, _ int) *int64 { return &service.Id })
	repositories, err := db.GetIssueRepositories(&entity.IssueRepositoryFilter{ServiceId: serviceIds})
	if err != nil {
		l.WithField("event-step", "FetchIssueRepositories").WithField("serviceIds", serviceIds).WithError(err).Error("Error while fetching issue repositories related to services that are related to the component instance")
		return nil, err
	}

	if len(repositories) < 1 {
		l.WithField("event-step", "FetchIssueRepositories").WithField("serviceIds", serviceIds).Error("No issue repositories found that are related to the services")
		return nil, NewIssueMatchHandlerError("No issue repositories found that are related to the services")
	}

	// Get Issue Variants
	// - getting high prio
	maxPriorityIr := lo.MaxBy(repositories, func(item entity.IssueRepository, max entity.IssueRepository) bool {
		return item.Priority > max.Priority
	})
	issueVariants, err := db.GetIssueVariants(&entity.IssueVariantFilter{
		IssueId:           lo.Map(issues, func(i entity.Issue, _ int) *int64 { return &i.Id }),
		IssueRepositoryId: []*int64{&maxPriorityIr.Id},
	})

	if err != nil {
		l.WithField("event-step", "FetchIssueVariants").WithError(err).Error("Error while fetching issue variants related to issue repositories")
		return nil, err
	}

	if len(issueVariants) < 1 {
		l.WithField("event-step", "FetchIssueVariants").Error("No issue variants found that are related to the issue repository")
		return nil, NewIssueMatchHandlerError("No issue variants found that are related to the issue repository")
	}

	// create a map of issue id to variants for easy access
	var issueVariantMap = make(map[int64]entity.IssueVariant)

	for _, variant := range issueVariants {
		// if there are multiple variants with the same priority on their repositories we take the highest severity one
		if _, ok := issueVariantMap[variant.IssueId]; ok {
			if issueVariantMap[variant.IssueId].Severity.Score < variant.Severity.Score {
				issueVariantMap[variant.IssueId] = variant
			}
		} else {
			issueVariantMap[variant.IssueId] = variant
		}
	}

	return issueVariantMap, nil
}

// OnComponentVersionAssignmentToComponentInstance is an event handler that is triggered when a component version is assigned to a component instance.
// It creates for the component instance ID that is assigned to component version ID a new issue match for each issue that is related to the component version.
// to do so it utilizes BuildIssueVariantMap to get the issue variants for the issues to identify the severity of the issueMatch.
func OnComponentVersionAssignmentToComponentInstance(db database.Database, componentInstanceID int64, componentVersionID int64) {
	l := logrus.WithFields(logrus.Fields{
		"event":               "IssueMatching.OnComponentVersionAssignmentToComponentInstance",
		"componentInstanceID": componentInstanceID,
		"componentVersionID":  componentVersionID,
	})

	l.WithField("event-step", "BuildIssueVariants").Debug("Building map of IssueVariants for issues related to assigned Component Version")
	issueVariantMap, err := BuildIssueVariantMap(db, componentInstanceID, componentVersionID)

	l.WithField("issueVariantMap", issueVariantMap)
	if err != nil {
		l.WithField("event-step", "BuildIssueVariants").WithError(err).Error("Error while fetching issues related to component Version")
		return
	}

	// For each issue create issue Match if not already exists
	for issueId, issueVariant := range issueVariantMap {
		l = l.WithFields(logrus.Fields{
			"issue": issueVariant,
		})

		l.WithField("event-step", "FetchIssueMatches").Debug("Fetching issue matches related to assigned Component Instance")
		issue_matches, err := db.GetIssueMatches(&entity.IssueMatchFilter{
			IssueId:             []*int64{&issueId},
			ComponentInstanceId: []*int64{&componentInstanceID},
		})
		if err != nil {
			l.WithField("event-step", "FetchIssueMatches").WithError(err).Error("Error while fetching issue matches related to assigned Component Instance")
			return
		}
		l.WithField("issueMatchesCount", len(issue_matches))

		// If the issue match is already created, we ignore it, as we do not associate with a version change a severity change
		if len(issue_matches) != 0 {
			l.WithField("event-step", "Skipping").Debug("The issue match does already exist. Skipping")
			return
		}

		// Create new issue match
		issue_match := &entity.IssueMatch{
			Status:                entity.IssueMatchStatusValuesNew,
			Severity:              issueVariantMap[issueId].Severity, //we got two  simply take the first one
			ComponentInstanceId:   componentInstanceID,
			IssueId:               issueId,
			TargetRemediationDate: GetTargetRemediationTimeline(issueVariant.Severity, time.Now()),
		}
		l.WithField("event-step", "CreateIssueMatch").WithField("issueMatch", issue_match).Debug("Creating Issue Match")

		_, err = db.CreateIssueMatch(issue_match)
		if err != nil {
			l.WithField("event-step", "CreateIssueMatch").WithError(err).Error("Error while creating issue match")
			return
		}
	}
}

func GetTargetRemediationTimeline(severity entity.Severity, creationDate time.Time) time.Time {
	//@todo get the configuration from environment variables or configuration file
	switch entity.SeverityValues(severity.Value) {
	case entity.SeverityValuesLow:
		return creationDate.AddDate(0, 6, 0)
	case entity.SeverityValuesMedium:

		return creationDate.AddDate(0, 3, 0)
	case entity.SeverityValuesHigh:
		return creationDate.AddDate(0, 0, 20)
	case entity.SeverityValuesCritical:
		return creationDate.AddDate(0, 0, 7)
	default:
		return time.Date(5000, 1, 1, 0, 0, 0, 0, time.UTC)
	}

}
