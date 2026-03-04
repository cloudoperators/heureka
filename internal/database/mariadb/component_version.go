// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"
	"strings"

	"github.com/cloudoperators/heureka/internal/database"
	"github.com/samber/lo"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

var componentVersionObject = DbObject{
	Properties: []*Property{
		NewProperty("componentversion_component_id", WrapChecker(func(cv *entity.ComponentVersion) bool { return cv.ComponentId != 0 })),
		NewProperty("componentversion_version", WrapChecker(func(cv *entity.ComponentVersion) bool { return cv.Version != "" })),
		NewProperty("componentversion_tag", WrapChecker(func(cv *entity.ComponentVersion) bool { return cv.Tag != "" })),
		NewProperty("componentversion_repository", WrapChecker(func(cv *entity.ComponentVersion) bool { return cv.Repository != "" })),
		NewProperty("componentversion_organization", WrapChecker(func(cv *entity.ComponentVersion) bool { return cv.Organization != "" })),
		NewImmutableProperty("componentversion_created_by"),
		NewProperty("componentversion_updated_by", WrapChecker(func(cv *entity.ComponentVersion) bool { return cv.UpdatedBy != 0 })),
		NewProperty("componentversion_end_of_life", WrapChecker(func(cv *entity.ComponentVersion) bool { return cv.EndOfLife != nil })),
	},
	FilterProperties: []*FilterProperty{
		NewFilterProperty("CV.componentversion_id = ?", WrapRetSlice(func(filter *entity.ComponentVersionFilter) []*int64 { return filter.Id })),
		NewFilterProperty("CVI.componentversionissue_issue_id = ?", WrapRetSlice(func(filter *entity.ComponentVersionFilter) []*int64 { return filter.IssueId })),
		NewFilterProperty("CV.componentversion_component_id = ?", WrapRetSlice(func(filter *entity.ComponentVersionFilter) []*int64 { return filter.ComponentId })),
		NewFilterProperty("CV.componentversion_version = ?", WrapRetSlice(func(filter *entity.ComponentVersionFilter) []*string { return filter.Version })),
		NewFilterProperty("CV.componentversion_tag = ?", WrapRetSlice(func(filter *entity.ComponentVersionFilter) []*string { return filter.Tag })),
		NewFilterProperty("CV.componentversion_repository = ?", WrapRetSlice(func(filter *entity.ComponentVersionFilter) []*string { return filter.Repository })),
		NewFilterProperty("CV.componentversion_organization = ?", WrapRetSlice(func(filter *entity.ComponentVersionFilter) []*string { return filter.Organization })),
		NewFilterProperty("C.component_ccrn = ?", WrapRetSlice(func(filter *entity.ComponentVersionFilter) []*string { return filter.ComponentCCRN })),
		NewFilterProperty("S.service_ccrn = ?", WrapRetSlice(func(filter *entity.ComponentVersionFilter) []*string { return filter.ServiceCCRN })),
		NewFilterProperty("CI.componentinstance_service_id = ?", WrapRetSlice(func(filter *entity.ComponentVersionFilter) []*int64 { return filter.ServiceId })),
		NewFilterProperty("IV.issuevariant_repository_id = ?", WrapRetSlice(func(filter *entity.ComponentVersionFilter) []*int64 { return filter.IssueRepositoryId })),
		NewFilterProperty("CV.componentversion_end_of_life = ?", WrapRetSlice(func(filter *entity.ComponentVersionFilter) []*bool { return filter.EndOfLife })),
		NewStateFilterProperty("CV.componentversion", WrapRetState(func(filter *entity.ComponentVersionFilter) []entity.StateFilterType { return filter.State })),
	},
}

func ensureComponentVersionFilter(filter *entity.ComponentVersionFilter) *entity.ComponentVersionFilter {
	if filter == nil {
		filter = &entity.ComponentVersionFilter{}
	}
	return EnsurePagination(filter)
}

func (s *SqlDatabase) getComponentVersionJoins(filter *entity.ComponentVersionFilter, order []entity.Order) string {
	joins := ""
	orderByCount := lo.ContainsBy(order, func(o entity.Order) bool {
		return o.By == entity.CriticalCount || o.By == entity.HighCount || o.By == entity.MediumCount || o.By == entity.LowCount || o.By == entity.NoneCount
	})
	if len(filter.IssueId) > 0 || orderByCount || len(filter.IssueRepositoryId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, "LEFT JOIN ComponentVersionIssue CVI on CV.componentversion_id = CVI.componentversionissue_component_version_id")
		if orderByCount || len(filter.IssueRepositoryId) > 0 {
			joins = fmt.Sprintf("%s\n%s", joins, "LEFT JOIN IssueVariant IV on IV.issuevariant_issue_id = CVI.componentversionissue_issue_id")
		}
	}
	if len(filter.ComponentCCRN) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, "LEFT JOIN Component C on CV.componentversion_component_id = C.component_id")
	}
	if len(filter.ServiceId) > 0 || len(filter.ServiceCCRN) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, "LEFT JOIN ComponentInstance CI on CV.componentversion_id = CI.componentinstance_component_version_id")

		if len(filter.ServiceCCRN) > 0 {
			joins = fmt.Sprintf("%s\n%s", joins, "LEFT JOIN Service S on S.service_id = CI.componentinstance_service_id")
		}
	}
	return joins
}

func (s *SqlDatabase) getComponentVersionColumns(order []entity.Order) string {
	columns := ""
	for _, o := range order {
		switch o.By {
		case entity.CriticalCount:
			columns = fmt.Sprintf("%s, COUNT(distinct CASE WHEN IV.issuevariant_rating = 'Critical' THEN IV.issuevariant_issue_id END) as critical_count", columns)
		case entity.HighCount:
			columns = fmt.Sprintf("%s, COUNT(distinct CASE WHEN IV.issuevariant_rating = 'High' THEN IV.issuevariant_issue_id END) as high_count", columns)
		case entity.MediumCount:
			columns = fmt.Sprintf("%s, COUNT(distinct CASE WHEN IV.issuevariant_rating = 'Medium' THEN IV.issuevariant_issue_id END) as medium_count", columns)
		case entity.LowCount:
			columns = fmt.Sprintf("%s, COUNT(distinct CASE WHEN IV.issuevariant_rating = 'Low' THEN IV.issuevariant_issue_id END) as low_count", columns)
		case entity.NoneCount:
			columns = fmt.Sprintf("%s, COUNT(distinct CASE WHEN IV.issuevariant_rating = 'None' THEN IV.issuevariant_issue_id END) as none_count", columns)
		}
	}
	return columns
}

func (s *SqlDatabase) buildComponentVersionStatement(baseQuery string, filter *entity.ComponentVersionFilter, withCursor bool, order []entity.Order, l *logrus.Entry) (Stmt, []interface{}, error) {
	filter = ensureComponentVersionFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	cursorFields, err := DecodeCursor(filter.Paginated.After)
	if err != nil {
		return nil, nil, err
	}
	cursorQuery := CreateCursorQuery("", cursorFields)

	order = GetDefaultOrder(order, entity.ComponentVersionId, entity.OrderDirectionAsc)
	orderStr := CreateOrderString(order)
	joins := s.getComponentVersionJoins(filter, order)

	filterStr := componentVersionObject.GetFilterQuery(filter)
	whereClause := ""
	if filterStr != "" {
		whereClause = fmt.Sprintf("WHERE %s", filterStr)
	}

	if withCursor && cursorQuery != "" {
		// cursor uses aggregated values that need to be used with having
		cursorQuery = fmt.Sprintf("HAVING (%s)", cursorQuery)
	}

	// construct final query
	var query string
	columns := s.getComponentVersionColumns(order)
	if withCursor {
		query = fmt.Sprintf(baseQuery, columns, joins, whereClause, cursorQuery, orderStr)
	} else {
		query = fmt.Sprintf(baseQuery, columns, joins, whereClause, orderStr)
	}

	// construct prepared statement and if where clause does exist add parameters
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

	filterParameters := componentVersionObject.GetFilterParameters(filter, withCursor, cursorFields)

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetAllComponentVersionCursors(filter *entity.ComponentVersionFilter, order []entity.Order) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetAllComponentVersionCursors",
	})

	baseQuery := `
		SELECT CV.* %s FROM ComponentVersion CV 
		%s
	    %s GROUP BY CV.componentversion_id ORDER BY %s
    `

	stmt, filterParameters, err := s.buildComponentVersionStatement(baseQuery, filter, false, order, l)
	if err != nil {
		return nil, err
	}

	rows, err := performListScan(
		stmt,
		filterParameters,
		l,
		func(l []RowComposite, e RowComposite) []RowComposite {
			return append(l, e)
		},
	)
	if err != nil {
		return nil, err
	}

	return lo.Map(rows, func(row RowComposite, _ int) string {
		cv := row.AsComponentVersion()
		var isc entity.IssueSeverityCounts
		if row.RatingCount != nil {
			isc = row.AsIssueSeverityCounts()
		}
		cursor, _ := EncodeCursor(WithComponentVersion(order, cv, isc))

		return cursor
	}), nil
}

func (s *SqlDatabase) GetComponentVersions(filter *entity.ComponentVersionFilter, order []entity.Order) ([]entity.ComponentVersionResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetComponentVersions",
	})

	baseQuery := `
		SELECT CV.* %s FROM ComponentVersion CV 
		%s
		%s
		GROUP BY CV.componentversion_id %s ORDER BY %s LIMIT ?
    `

	filter = ensureComponentVersionFilter(filter)

	stmt, filterParameters, err := s.buildComponentVersionStatement(baseQuery, filter, true, order, l)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.ComponentVersionResult, e RowComposite) []entity.ComponentVersionResult {
			cv := e.AsComponentVersion()

			var isc entity.IssueSeverityCounts
			if e.RatingCount != nil {
				isc = e.AsIssueSeverityCounts()
			}

			cursor, _ := EncodeCursor((WithComponentVersion(order, cv, isc)))

			cvr := entity.ComponentVersionResult{
				WithCursor: entity.WithCursor{
					Value: cursor,
				},
				ComponentVersion: &cv,
			}
			return append(l, cvr)
		},
	)
}

func (s *SqlDatabase) CountComponentVersions(filter *entity.ComponentVersionFilter) (int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountComponentVersions",
	})

	baseQuery := `
		SELECT count(distinct CV.componentversion_id) %s FROM ComponentVersion CV 
		%s
		%s
		ORDER BY %s
	`
	stmt, filterParameters, err := s.buildComponentVersionStatement(baseQuery, filter, false, []entity.Order{}, l)
	if err != nil {
		return -1, err
	}

	defer stmt.Close()

	return performCountScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) CreateComponentVersion(componentVersion *entity.ComponentVersion) (*entity.ComponentVersion, error) {
	l := logrus.WithFields(logrus.Fields{
		"componentVersion": componentVersion,
		"event":            "database.CreateComponentVersion",
	})

	componentVersionRow := ComponentVersionRow{}
	componentVersionRow.FromComponentVersion(componentVersion)

	query := componentVersionObject.InsertQuery("ComponentVersion")
	id, err := performInsert(s, query, componentVersionRow, l)
	if err != nil {
		if strings.HasPrefix(err.Error(), "Error 1062") {
			return nil, database.NewDuplicateEntryDatabaseError(fmt.Sprintf("for ComponentVersion: %s ", componentVersion.Version))
		}
		return nil, err
	}

	componentVersion.Id = id
	return componentVersion, nil
}

func (s *SqlDatabase) UpdateComponentVersion(componentVersion *entity.ComponentVersion) error {
	l := logrus.WithFields(logrus.Fields{
		"componentVersion": componentVersion,
		"event":            "database.UpdateComponentVersion",
	})

	baseQuery := `
		UPDATE ComponentVersion SET
		%s
		WHERE componentversion_id = :componentversion_id
	`

	updateFields := componentVersionObject.GetUpdateFields(componentVersion)

	query := fmt.Sprintf(baseQuery, updateFields)

	componentVersionRow := ComponentVersionRow{}
	componentVersionRow.FromComponentVersion(componentVersion)

	_, err := performExec(s, query, componentVersionRow, l)

	return err
}

func (s *SqlDatabase) DeleteComponentVersion(id int64, userId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": "database.DeleteComponentVersion",
	})

	query := `
		UPDATE ComponentVersion SET
		componentversion_deleted_at = NOW(),
		componentversion_updated_by = :userId
		WHERE componentversion_id = :id
	`

	args := map[string]interface{}{
		"userId": userId,
		"id":     id,
	}

	_, err := performExec(s, query, args, l)

	return err
}
