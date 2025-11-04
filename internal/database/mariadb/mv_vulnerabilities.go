// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

func getMvCountIssueRatingsJoin(filter *entity.IssueFilter) string {
	if filter.AllServices && filter.Unique {
		// Conunt unique issues. AllServices filter is set, so we count issues that are matched to a service
		// COUNT(distinct IV.issuevariant_issue_id)
		//	LEFT JOIN mvCountIssueRatingsUniqueService CIR ON IV.issuevariant_rating = CIR.issue_value
		return `
			LEFT JOIN (
				SELECT 'Critical' AS issue_value, critical_count AS issue_count FROM mvCountIssueRatingsUniqueService
				UNION ALL
				SELECT 'High', high_count FROM mvCountIssueRatingsUniqueService
				UNION ALL
				SELECT 'Medium', medium_count FROM mvCountIssueRatingsUniqueService
				UNION ALL
				SELECT 'Low', low_count FROM mvCountIssueRatingsUniqueService
				UNION ALL
				SELECT 'None', none_count FROM mvCountIssueRatingsUniqueService
			) CIR ON IV.issuevariant_rating = CIR.issue_value
		`
	} else if filter.AllServices {
		// Count issues that appear in multiple services and in multiple component versions per service
		//COUNT(distinct CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id))
		//		LEFT JOIN mvCountIssueRatingsService CIR ON SG.supportgroup_ccrn = CIR.supportgroup_ccrn
		//			AND IV.issuevariant_rating = CIR.issue_value
		if len(filter.SupportGroupCCRN) > 0 {
			return `
				LEFT JOIN (
					SELECT supportgroup_ccrn, critical_count AS issue_count, 'Critical' AS issue_value FROM mvCountIssueRatingsService
					UNION ALL
					SELECT supportgroup_ccrn, high_count, 'High' FROM mvCountIssueRatingsService
					UNION ALL
					SELECT supportgroup_ccrn, medium_count, 'Medium' FROM mvCountIssueRatingsService
					UNION ALL
					SELECT supportgroup_ccrn, low_count, 'Low' FROM mvCountIssueRatingsService
					UNION ALL
					SELECT supportgroup_ccrn, none_count, 'None' FROM mvCountIssueRatingsService
				) CIR
				ON SG.supportgroup_ccrn = CIR.supportgroup_ccrn
				AND IV.issuevariant_rating = CIR.issue_value
			`
		} else {
			// call/branch can be replaced with (something to consider):
			// SELECT issue_value, issue_count
			// FROM mvCountIssueRatingsServiceWithoutSupportGroup
			// ORDER BY issue_value ASC;
			//	LEFT JOIN mvCountIssueRatingsServiceWithoutSupportGroup CIR ON IV.issuevariant_rating = CIR.issue_value
			return `
				LEFT JOIN (
					SELECT critical_count AS issue_count, 'Critical' AS issue_value FROM mvCountIssueRatingsServiceWithoutSupportGroup
					UNION ALL
					SELECT high_count, 'High' FROM mvCountIssueRatingsServiceWithoutSupportGroup
					UNION ALL
					SELECT medium_count, 'Medium' FROM mvCountIssueRatingsServiceWithoutSupportGroup
					UNION ALL
					SELECT low_count, 'Low' FROM mvCountIssueRatingsServiceWithoutSupportGroup
					UNION ALL
					SELECT none_count, 'None' FROM mvCountIssueRatingsServiceWithoutSupportGroup
				) CIR
				ON IV.issuevariant_rating = CIR.issue_value
			`
		}
	} else if len(filter.SupportGroupCCRN) > 0 {
		// Count issues that appear in multiple support groups
		// COUNT(distinct CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', SGS.supportgroupservice_service_id, ',', SG.supportgroup_id))
		//	LEFT JOIN mvCountIssueRatingsSupportGroup CIR ON SG.supportgroup_ccrn = CIR.supportgroup_ccrn
		//		AND IV.issuevariant_rating = CIR.issue_value
		return `
			LEFT JOIN (
				SELECT supportgroup_ccrn, critical_count AS issue_count, 'Critical' AS issue_value FROM mvCountIssueRatingsSupportGroup
				UNION ALL
				SELECT supportgroup_ccrn, high_count, 'High' FROM mvCountIssueRatingsSupportGroup
				UNION ALL
				SELECT supportgroup_ccrn, medium_count, 'Medium' FROM mvCountIssueRatingsSupportGroup
				UNION ALL
				SELECT supportgroup_ccrn, low_count, 'Low' FROM mvCountIssueRatingsSupportGroup
				UNION ALL
				SELECT supportgroup_ccrn, none_count, 'None' FROM mvCountIssueRatingsSupportGroup
			) CIR
			ON SG.supportgroup_ccrn = CIR.supportgroup_ccrn
			AND IV.issuevariant_rating = CIR.issue_value
		`
	} else if len(filter.ComponentVersionId) > 0 {
		// Count issues that appear in multiple component versions
		// COUNT(DISTINCT CONCAT(CVI.componentversionissue_component_version_id, ',', CVI.componentversionissue_issue_id)) "
		//	LEFT JOIN mvCountIssueRatingsComponentVersion CIR ON CVI.componentversionissue_component_version_id = CIR.component_version_id
		//		AND IV.issuevariant_rating = CIR.issue_value
		return `
			LEFT JOIN (
				SELECT component_version_id, critical_count AS issue_count, 'Critical' AS issue_value FROM mvCountIssueRatingsComponentVersion
				UNION ALL
				SELECT component_version_id, high_count, 'High' FROM mvCountIssueRatingsComponentVersion
				UNION ALL
				SELECT component_version_id, medium_count, 'Medium' FROM mvCountIssueRatingsComponentVersion
				UNION ALL
				SELECT component_version_id, low_count, 'Low' FROM mvCountIssueRatingsComponentVersion
				UNION ALL
				SELECT component_version_id, none_count, 'None' FROM mvCountIssueRatingsComponentVersion
			) CIR
			ON CVI.componentversionissue_component_version_id = CIR.component_version_id
			AND IV.issuevariant_rating = CIR.issue_value
		`
	} else if len(filter.ServiceCCRN) > 0 || len(filter.ServiceId) > 0 {
		// COUNT(distinct CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id))
		//	LEFT JOIN mvCountIssueRatingsServiceId CIR ON CI.componentinstance_service_id = CIR.service_id
		//		AND IV.issuevariant_rating = CIR.issue_value
		return `
			LEFT JOIN (
				SELECT service_id, critical_count AS issue_count, 'Critical' AS issue_value FROM mvCountIssueRatingsServiceId
				UNION ALL
				SELECT service_id, high_count, 'High' FROM mvCountIssueRatingsServiceId
				UNION ALL
				SELECT service_id, medium_count, 'Medium' FROM mvCountIssueRatingsServiceId
				UNION ALL
				SELECT service_id, low_count, 'Low' FROM mvCountIssueRatingsServiceId
				UNION ALL
				SELECT service_id, none_count, 'None' FROM mvCountIssueRatingsServiceId
			) CIR
			ON CI.componentinstance_service_id = CIR.service_id
			AND IV.issuevariant_rating = CIR.issue_value
		`
	} else {
		// COUNT(distinct IV.issuevariant_issue_id)
		//	LEFT JOIN mvCountIssueRatingsOther CIR ON IV.issuevariant_rating = CIR.issue_value
		return `
			LEFT JOIN (
				SELECT critical_count AS issue_count, 'Critical' AS issue_value FROM mvCountIssueRatingsOther
				UNION ALL
				SELECT high_count, 'High' FROM mvCountIssueRatingsOther
				UNION ALL
				SELECT medium_count, 'Medium' FROM mvCountIssueRatingsOther
				UNION ALL
				SELECT low_count, 'Low' FROM mvCountIssueRatingsOther
				UNION ALL
				SELECT none_count, 'None' FROM mvCountIssueRatingsOther
			) CIR
			ON IV.issuevariant_rating = CIR.issue_value
		`
	}
}

func getIssueJoinsWithMvCountIssueRatingsJoin(filter *entity.IssueFilter, order []entity.Order) string {
	joins := getIssueJoins(filter, order)
	joins = fmt.Sprintf("%s\n%s", joins, getMvCountIssueRatingsJoin(filter))
	return joins
}

func getIssueQueryWithMvCountIssueRatingsJoin(baseQuery string, order []entity.Order, filter *entity.IssueFilter) string {
	issueColumns := getIssueColumns(order)
	defaultOrder := GetDefaultOrder(order, entity.IssueId, entity.OrderDirectionAsc)
	joins := getIssueJoinsWithMvCountIssueRatingsJoin(filter, order)
	whereClause := getIssueFilterWhereClause(filter)
	orderStr := CreateOrderString(defaultOrder)
	return fmt.Sprintf(baseQuery, issueColumns, joins, whereClause, orderStr)
}

func (s *SqlDatabase) buildIssueStatementWithMvCountIssueRatingsJoin(baseQuery string, filter *entity.IssueFilter, order []entity.Order, l *logrus.Entry) (Stmt, []interface{}, error) {
	ifilter := s.ensureIssueFilter(filter)
	l.WithFields(logrus.Fields{"filter": ifilter})

	cursorFields, err := DecodeCursor(ifilter.PaginatedX.After)
	if err != nil {
		return nil, nil, err
	}

	query := getIssueQueryWithMvCountIssueRatingsJoin(baseQuery, order, ifilter)

	//construct prepared statement and if where clause does exist add parameters
	stmt, err := s.db.Preparex(query)
	if err != nil {
		msg := ERROR_MSG_PREPARED_STMT
		l.WithFields(
			logrus.Fields{
				"error": err,
				"query": query,
				"stmt":  stmt,
			}).Error(msg)
		return nil, nil, fmt.Errorf("%s", msg)
	}

	//adding parameters
	filterParameters := s.buildIssueFilterParameters(ifilter, cursorFields)

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) CountIssueRatings(filter *entity.IssueFilter) (*entity.IssueSeverityCounts, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountIssueRatings",
	})

	filter = s.ensureIssueFilter(filter)

	baseQuery := `
		SELECT IV.issuevariant_rating AS issue_value, CIR.issue_count AS issue_count FROM %s Issue I
		%s
		%s
		%s
		GROUP BY IV.issuevariant_rating ORDER BY %s
	`

	if len(filter.IssueRepositoryId) == 0 {
		baseQuery = fmt.Sprintf(baseQuery, "%s", "LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id", "%s", "%s", "%s")
	}

	stmt, filterParameters, err := s.buildIssueStatementWithMvCountIssueRatingsJoin(baseQuery, filter, []entity.Order{}, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	counts, err := performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.IssueCount, e IssueCountRow) []entity.IssueCount {
			return append(l, e.AsIssueCount())
		},
	)

	if err != nil {
		return nil, err
	}

	var issueSeverityCounts entity.IssueSeverityCounts
	for _, count := range counts {
		switch count.Value {
		case entity.SeverityValuesCritical.String():
			issueSeverityCounts.Critical = count.Count
		case entity.SeverityValuesHigh.String():
			issueSeverityCounts.High = count.Count
		case entity.SeverityValuesMedium.String():
			issueSeverityCounts.Medium = count.Count
		case entity.SeverityValuesLow.String():
			issueSeverityCounts.Low = count.Count
		case entity.SeverityValuesNone.String():
			issueSeverityCounts.None = count.Count
		}
		issueSeverityCounts.Total += count.Count
	}
	return &issueSeverityCounts, nil
}
