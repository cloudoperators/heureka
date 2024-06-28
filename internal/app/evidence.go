// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

func (h *HeurekaApp) getEvidenceResults(filter *entity.EvidenceFilter) ([]entity.EvidenceResult, error) {
	var evidenceResults []entity.EvidenceResult
	evidences, err := h.database.GetEvidences(filter)
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

func (h *HeurekaApp) ListEvidences(filter *entity.EvidenceFilter, options *entity.ListOptions) (*entity.List[entity.EvidenceResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	ensurePaginated(&filter.Paginated)

	l := logrus.WithFields(logrus.Fields{
		"event":  "app.ListEvidences",
		"filter": filter,
	})

	res, err := h.getEvidenceResults(filter)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Error while filtering for Evidences")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := h.database.GetAllEvidenceIds(filter)
			if err != nil {
				l.Error(err)
				return nil, heurekaError("Error while getting all Ids")
			}
			pageInfo = getPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = h.database.CountEvidences(filter)
		if err != nil {
			l.Error(err)
			return nil, heurekaError("Error while total count of Evidences")
		}
	}

	return &entity.List[entity.EvidenceResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}, nil
}

func (h *HeurekaApp) CreateEvidence(evidence *entity.Evidence) (*entity.Evidence, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  "app.CreateEvidence",
		"object": evidence,
	})

	newEvidence, err := h.database.CreateEvidence(evidence)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while creating evidence.")
	}

	return newEvidence, nil
}

func (h *HeurekaApp) UpdateEvidence(evidence *entity.Evidence) (*entity.Evidence, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  "app.UpdateEvidence",
		"object": evidence,
	})

	err := h.database.UpdateEvidence(evidence)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while updating evidence.")
	}

	evidenceResult, err := h.ListEvidences(&entity.EvidenceFilter{Id: []*int64{&evidence.Id}}, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while retrieving updated evidence.")
	}

	if len(evidenceResult.Elements) != 1 {
		l.Error(err)
		return nil, heurekaError("Multiple evidences found.")
	}

	return evidenceResult.Elements[0].Evidence, nil
}

func (h *HeurekaApp) DeleteEvidence(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": "app.DeleteEvidence",
		"id":    id,
	})

	err := h.database.DeleteEvidence(id)

	if err != nil {
		l.Error(err)
		return heurekaError("Internal error while deleting evidence.")
	}

	return nil
}
