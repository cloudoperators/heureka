// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_match

import (
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/app/component_instance"
	"github.wdf.sap.corp/cc/heureka/internal/app/event"
	"github.wdf.sap.corp/cc/heureka/internal/database"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
	"github.wdf.sap.corp/cc/heureka/pkg/util"
	"time"
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

// @todo method is to long and needs a refactoring.
func OnComponentVersionAssignmentToComponentInstance(db database.Database, componentInstanceID int64, componentVersionID int64) {
	l := logrus.WithFields(logrus.Fields{
		"event":               "IssueMatching.OnComponentVersionAssignmentToComponentInstance",
		"componentInstanceID": componentInstanceID,
		"componentVersionID":  componentVersionID,
	})

	l.WithField("event-step", "FetchIssues").Debug("Fetching issues related to assigned Component Version")

	issues, err := db.GetIssues(&entity.IssueFilter{
		Paginated:          entity.Paginated{},
		ComponentVersionId: []*int64{&componentVersionID},
	})

	if err != nil {
		l.WithField("event-step", "FetchIssues").WithError(err).Error("Error while fetching issues related to component Version")
		return
	}

	for _, issue := range issues {
		l = l.WithFields(logrus.Fields{
			"issue": issue,
		})

		l.WithField("event-step", "FetchIssueMatches").Debug("Fetching issue matches related to assigned Component Instance")
		issue_matches, err := db.GetIssueMatches(&entity.IssueMatchFilter{
			Paginated:           entity.Paginated{First: util.Ptr(1000000)},
			IssueId:             []*int64{&issue.Id},
			ComponentInstanceId: []*int64{&componentInstanceID},
		})

		if err != nil {
			l.WithField("event-step", "FetchIssueMatches").WithError(err).Error("Error while fetching issue matches related to assigned Component Instance")
			return
		}

		//if the match already created we ignore it as we do not associate with a version change a severity change
		if len(issue_matches) != 0 {
			l.WithField("event-step", "Skipping").Debug("The issue match does already exist. Skipping")
			return
		}

		l.WithField("event-step", "FetchServices").Debug("Fetching Services related to Component Instance")
		services, err := db.GetServices(&entity.ServiceFilter{
			Paginated:           entity.Paginated{},
			ComponentInstanceId: []*int64{&componentInstanceID},
		})

		if err != nil {
			l.WithField("event-step", "FetchServices").WithError(err).Error("Error while fetching services related to Component Instance")
			return
		}

		serviceIds := lo.Map(services, func(service entity.Service, _ int) *int64 {
			return &service.Id
		})

		l.WithField("event-step", "FetchIssueRepositories").WithField("serviceIds", serviceIds).Debug("Fetching Issue Repositories related to Services that are related to the Component Instance")
		repositories, err := db.GetIssueRepositories(&entity.IssueRepositoryFilter{
			Paginated: entity.Paginated{},
			ServiceId: serviceIds,
		})

		if err != nil {
			l.WithField("event-step", "FetchIssueRepositories").WithField("serviceIds", serviceIds).WithError(err).Error("Error while fetching issue repositories related to services that are related to the component instance")
			return
		}

		if len(repositories) < 1 {
			l.WithField("event-step", "FetchIssueRepositories").WithField("serviceIds", serviceIds).Error("No issue repositories found that are related to the services")
			return
		}

		l.WithField("event-step", "FetchIssueVariants").Debug("Identifying highes prio Issue Repository and fetching Issue Variants related to it")
		//getting high prio
		maxPriorityIr := lo.MaxBy(repositories, func(item entity.IssueRepository, max entity.IssueRepository) bool {
			return item.Priority > max.Priority
		})

		maxPrioRepoIds := lo.FilterMap(repositories, func(item entity.IssueRepository, index int) (*int64, bool) {
			if item.Priority == maxPriorityIr.Priority {
				return &item.Id, true
			}
			return util.Ptr(int64(0)), false
		})

		issueIds := lo.Map(issues, func(i entity.Issue, _ int) *int64 { return &i.Id })

		variants, err := db.GetIssueVariants(&entity.IssueVariantFilter{
			Paginated:         entity.Paginated{},
			IssueId:           issueIds,
			IssueRepositoryId: maxPrioRepoIds,
		})

		if err != nil {
			l.WithField("event-step", "FetchIssueVariants").WithError(err).Error("Error while fetching issue variants related to issue repositories")
			return
		}

		if len(variants) < 1 {
			l.WithField("event-step", "FetchIssueVariants").Error("No issue variants found that are related to the issue repository")
			return
		}

		issue_match := &entity.IssueMatch{
			Status:                entity.IssueMatchStatusValuesNew,
			Severity:              variants[0].Severity,
			ComponentInstanceId:   componentInstanceID,
			IssueId:               issue.Id,
			TargetRemediationDate: GetTargetRemediationTimeline(variants[0].Severity, time.Now()),
		}
		l.WithField("event-step", "CreateIssueMatch").WithField("issueMatch", issue_match).Debug("Creating Issue Match")

		_, err = db.CreateIssueMatch(issue_match)

		if err != nil {
			l.WithField("event-step", "CreateIssueMatch").WithError(err).Error("Error while creating issue match")
			return
		}

	}
}

//
//func ComponentVersionDelta(db database.Database, oldComponentVersionID int64, newComponentVersionID int64) ([]*int64, []*int64) {
//
//	oldIssues, err := db.GetIssues(&entity.IssueFilter{
//		Paginated:          entity.Paginated{},
//		ComponentVersionId: []*int64{&oldComponentVersionID},
//	})
//
//	if err != nil {
//		//log error
//		return make([]*int64, 0), make([]*int64, 0)
//	}
//	newIssues, err := db.GetIssues(&entity.IssueFilter{
//		Paginated:          entity.Paginated{},
//		ComponentVersionId: []*int64{&newComponentVersionID},
//	})
//
//	if err != nil {
//		//log error
//		return make([]*int64, 0), make([]*int64, 0)
//	}
//
//	oldIssueIds := lo.Map(oldIssues, func(i entity.Issue, _ int) *int64 { return &i.Id })
//	newIssueIds := lo.Map(newIssues, func(i entity.Issue, _ int) *int64 { return &i.Id })
//
//	added, removed := lo.Difference(oldIssueIds, newIssueIds)
//
//	return added, removed
//}

func GetTargetRemediationTimeline(severity entity.Severity, creationDate time.Time) time.Time {
	//@todo get the configuration from environment variables
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
