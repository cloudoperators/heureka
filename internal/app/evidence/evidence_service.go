// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package evidence

import (
	"fmt"
	"github.wdf.sap.corp/cc/heureka/internal/app/common"
	"github.wdf.sap.corp/cc/heureka/internal/app/event"
	"github.wdf.sap.corp/cc/heureka/internal/database"

	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

type evidenceService struct {
	database      database.Database
	eventRegistry event.EventRegistry
}

func NewEvidenceService(db database.Database, er event.EventRegistry) EvidenceService {
	return &evidenceService{
		database:      db,
		eventRegistry: er,
	}
}

type EvidenceServiceError struct {
	msg string
}

func (e *EvidenceServiceError) Error() string {
	return fmt.Sprintf("EvidenceServiceError: %s", e.msg)
}

func NewEvidenceServiceError(msg string) *EvidenceServiceError {
	return &EvidenceServiceError{msg: msg}
}

func (e *evidenceService) getEvidenceResults(filter *entity.EvidenceFilter) ([]entity.EvidenceResult, error) {
	var evidenceResults []entity.EvidenceResult
	evidences, err := e.database.GetEvidences(filter)
	if err != nil {
		return nil, err
	}
	for _, e := range evidences {
		evidence := e
		cursor := fmt.Sprintf("%d", evidence.Id)
		evidenceResults = append(evidenceResults, entity.EvidenceResult{
			WithCursor:           entity.WithCursor{Value: cursor},
			EvidenceAggregations: nil,
			Evidence:             &evidence,
		})
	}
	return evidenceResults, nil
}

func (e *evidenceService) ListEvidences(filter *entity.EvidenceFilter, options *entity.ListOptions) (*entity.List[entity.EvidenceResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	common.EnsurePaginated(&filter.Paginated)

	l := logrus.WithFields(logrus.Fields{
		"event":  ListEvidencesEventName,
		"filter": filter,
	})

	res, err := e.getEvidenceResults(filter)

	if err != nil {
		l.Error(err)
		return nil, NewEvidenceServiceError("Error while filtering for Evidences")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := e.database.GetAllEvidenceIds(filter)
			if err != nil {
				l.Error(err)
				return nil, NewEvidenceServiceError("Error while getting all Ids")
			}
			pageInfo = common.GetPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = e.database.CountEvidences(filter)
		if err != nil {
			l.Error(err)
			return nil, NewEvidenceServiceError("Error while total count of Evidences")
		}
	}

	ret := &entity.List[entity.EvidenceResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}

	e.eventRegistry.PushEvent(&ListEvidencesEvent{Filter: filter, Options: options, Results: ret})

	return ret, nil
}

func (e *evidenceService) CreateEvidence(evidence *entity.Evidence) (*entity.Evidence, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  CreateEvidenceEventName,
		"object": evidence,
	})

	newEvidence, err := e.database.CreateEvidence(evidence)

	if err != nil {
		l.Error(err)
		return nil, NewEvidenceServiceError("Internal error while creating evidence.")
	}

	e.eventRegistry.PushEvent(&CreateEvidenceEvent{Evidence: newEvidence})

	return newEvidence, nil
}

func (e *evidenceService) UpdateEvidence(evidence *entity.Evidence) (*entity.Evidence, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  UpdateEvidenceEventName,
		"object": evidence,
	})

	err := e.database.UpdateEvidence(evidence)

	if err != nil {
		l.Error(err)
		return nil, NewEvidenceServiceError("Internal error while updating evidence.")
	}

	evidenceResult, err := e.ListEvidences(&entity.EvidenceFilter{Id: []*int64{&evidence.Id}}, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, NewEvidenceServiceError("Internal error while retrieving updated evidence.")
	}

	if len(evidenceResult.Elements) != 1 {
		l.Error(err)
		return nil, NewEvidenceServiceError("Multiple evidences found.")
	}

	e.eventRegistry.PushEvent(&UpdateEvidenceEvent{Evidence: evidence})

	return evidenceResult.Elements[0].Evidence, nil
}

func (e *evidenceService) DeleteEvidence(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": DeleteEvidenceEventName,
		"id":    id,
	})

	err := e.database.DeleteEvidence(id)

	if err != nil {
		l.Error(err)
		return NewEvidenceServiceError("Internal error while deleting evidence.")
	}

	e.eventRegistry.PushEvent(&DeleteEvidenceEvent{EvidenceID: id})

	return nil
}
