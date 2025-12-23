// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

func getCountTable(filter *entity.IssueFilter) string {
	if filter.AllServices && filter.Unique {
		// Total count of unique issues
		return "mvCountIssueRatingsUniqueService"
	} else if filter.AllServices {
		if len(filter.SupportGroupCCRN) > 0 {
			// total count of issues in support group across all services
			// service list view total count with support group filter
			return "mvCountIssueRatingsService"
		} else {
			// total count of issues in all services (across all support groups)
			// service list view total count without support group filter
			return "mvCountIssueRatingsServiceWithoutSupportGroup"
		}
	} else if len(filter.SupportGroupCCRN) > 0 && len(filter.ServiceCCRN) == 0 && len(filter.ServiceId) == 0 {
		// Count issues in a support group
		return "mvCountIssueRatingsSupportGroup"
	} else if len(filter.ComponentVersionId) > 0 {
		// Count issues in a component version of a *service*
		return "mvCountIssueRatingsComponentVersion"
	} else if len(filter.ServiceCCRN) > 0 || len(filter.ServiceId) > 0 {
		// Count issues that appear in single service
		return "mvCountIssueRatingsServiceId"
	} else {
		// Total count of issues
		return "mvCountIssueRatingsOther"
	}
}

func (s *SqlDatabase) CountIssueRatings(filter *entity.IssueFilter) (*entity.IssueSeverityCounts, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountIssueRatings",
	})
	var fl []string
	var filterParameters []any

	filter = s.ensureIssueFilter(filter)

	baseQuery := `
		SELECT CIR.critical_count, CIR.high_count, CIR.medium_count, CIR.low_count, CIR.none_count FROM %s AS CIR
	`

	tableName := getCountTable(filter)

	query := fmt.Sprintf(baseQuery, tableName)

	if len(filter.ServiceId) > 0 {
		filterParameters = buildQueryParameters(filterParameters, filter.ServiceId)
		fl = append(fl, buildFilterQuery(filter.ServiceId, "CIR.service_id = ?", OP_OR))
	}

	if len(filter.ServiceCCRN) > 0 {
		filterParameters = buildQueryParameters(filterParameters, filter.ServiceCCRN)
		fl = append(fl, buildFilterQuery(filter.ServiceCCRN, "CIR.service_ccrn = ?", OP_OR))
	}

	if len(filter.ComponentVersionId) > 0 {
		filterParameters = buildQueryParameters(filterParameters, filter.ComponentVersionId)
		fl = append(fl, buildFilterQuery(filter.ComponentVersionId, "CIR.component_version_id = ?", OP_OR))
	}

	if len(filter.SupportGroupCCRN) > 0 && len(filter.ServiceId) == 0 && len(filter.ServiceCCRN) == 0 {
		filterParameters = buildQueryParameters(filterParameters, filter.SupportGroupCCRN)
		fl = append(fl, buildFilterQuery(filter.SupportGroupCCRN, "CIR.supportgroup_ccrn = ?", OP_OR))
	}

	filterStr := combineFilterQueries(fl, OP_AND)
	if filterStr != "" {
		query = fmt.Sprintf("%s WHERE %s", query, filterStr)
	}

	stmt, err := s.db.Preparex(query)
	if err != nil {
		msg := ERROR_MSG_PREPARED_STMT
		l.WithFields(
			logrus.Fields{
				"error": err,
				"query": query,
				"stmt":  stmt,
			}).Error(msg)
		return nil, fmt.Errorf("%s", msg)
	}

	defer stmt.Close()

	counts, err := performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.IssueSeverityCounts, e RatingCount) []entity.IssueSeverityCounts {
			return append(l, e.AsIssueSeverityCounts())
		},
	)

	if err != nil {
		return nil, err
	}

	if len(counts) == 0 {
		return &entity.IssueSeverityCounts{
			Critical: 0,
			High:     0,
			Medium:   0,
			Low:      0,
			None:     0,
		}, nil
	}

	return &counts[0], err
}
