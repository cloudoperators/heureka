// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_match

import (
	"time"

	"github.com/cloudoperators/heureka/internal/app/component_instance"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
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
	ListIssueMatchIDsEventName            event.EventName = "ListIssueMatchIDs"
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
// it does take the first issue_variant with the highest priority for the respective component instance.
// This is archived by utilizing database.GetServiceIssueVariants that does return ALL issue variants for a given
// component instance id together with the priorty and afterwards identifying for each issue the variant with the highest
// priority
//
// Returns a map of issue id to issue variant
func BuildIssueVariantMap(db database.Database, componentInstanceID int64, componentVersionID int64) (map[int64]entity.ServiceIssueVariant, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":               "BuildIssueVariantMap",
		"componentInstanceID": componentInstanceID,
		"componentVersionID":  componentVersionID,
	})

	// Get Issue Variants
	issueVariants, err := db.GetServiceIssueVariants(&entity.ServiceIssueVariantFilter{ComponentInstanceId: []*int64{&componentInstanceID}})
	if err != nil {
		l.WithField("event-step", "FetchIssueVariants").WithError(err).Error("Error while fetching issue variants")
		return nil, NewIssueMatchHandlerError("Error while fetching issue variants")
	}

	//No issue variants found,
	if len(issueVariants) < 1 {
		l.WithField("event-step", "FetchIssueVariants").Error("No issue variants found that are related to the issue repository")
		return nil, NewIssueMatchHandlerError("No issue variants found that are related to the issue repository")
	}

	// create a map of issue id to variants for easy access
	var issueVariantMap = make(map[int64]entity.ServiceIssueVariant)

	for _, variant := range issueVariants {

		if _, ok := issueVariantMap[variant.IssueId]; ok {
			// if there are multiple variants with the same priority on their repositories we take the highest severity one
			// if serverity and score are the same the first occuring issue variant is taken
			if issueVariantMap[variant.IssueId].Priority < variant.Priority {
				issueVariantMap[variant.IssueId] = variant
			} else if issueVariantMap[variant.IssueId].Severity.Score < variant.Severity.Score {
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
		l.WithField("event-step", "BuildIssueVariants").WithError(err).Error("Error while fetching issues related to component instance")
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
		// currently a static user is assumed to be used, this going to change in future to either a configured user or a dynamically
		// infered user from the component version issue macht
		issue_match := &entity.IssueMatch{
			UserId:                1, //@todo discuss whatever we use a static system user or infer the user from the ComponentVersionIssue
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

// GetTargetRemediationTimeline returns the target remediation timeline for a given severity and creation date
// In a first iteration this is going to be obtained from a static configuration in environment variables or configuration file
// In future this is potentially going to be dynamically inferred from individual service configurations in addition
//
// @todo get the configuration from environment variables or configuration file
func GetTargetRemediationTimeline(severity entity.Severity, creationDate time.Time) time.Time {
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
