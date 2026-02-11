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

func ensureComponentVersionFilter(f *entity.ComponentVersionFilter) *entity.ComponentVersionFilter {
	first := 1000
	after := ""
	if f == nil {
		return &entity.ComponentVersionFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
			Id:            nil,
			IssueId:       nil,
			ComponentCCRN: nil,
			ComponentId:   nil,
			Tag:           nil,
			Repository:    nil,
			Organization:  nil,
		}
	}
	if f.First == nil {
		f.First = &first
	}
	if f.After == nil {
		f.After = &after
	}
	return f
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

func getComponentVersionUpdateFields(componentVersion *entity.ComponentVersion) string {
	fl := []string{}
	if componentVersion.Version != "" {
		fl = append(fl, "componentversion_version = :componentversion_version")
	}
	if componentVersion.ComponentId != 0 {
		fl = append(fl, "componentversion_component_id = :componentversion_component_id")
	}
	if componentVersion.Tag != "" {
		fl = append(fl, "componentversion_tag = :componentversion_tag")
	}
	if componentVersion.Repository != "" {
		fl = append(fl, "componentversion_repository = :componentversion_repository")
	}
	if componentVersion.Organization != "" {
		fl = append(fl, "componentversion_organization = :componentversion_organization")
	}
	if componentVersion.UpdatedBy != 0 {
		fl = append(fl, "componentversion_updated_by = :componentversion_updated_by")
	}
	return strings.Join(fl, ", ")
}

func getComponentVersionFilterString(filter *entity.ComponentVersionFilter) string {
	var fl []string
	fl = append(fl, buildFilterQuery(filter.Id, "CV.componentversion_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueId, "CVI.componentversionissue_issue_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ComponentId, "CV.componentversion_component_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Version, "CV.componentversion_version = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Tag, "CV.componentversion_tag = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Repository, "CV.componentversion_repository = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Organization, "CV.componentversion_organization = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ComponentCCRN, "C.component_ccrn = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ServiceCCRN, "S.service_ccrn = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ServiceId, "CI.componentinstance_service_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueRepositoryId, "IV.issuevariant_repository_id = ?", OP_OR))
	fl = append(fl, buildStateFilterQuery(filter.State, "CV.componentversion"))

	return combineFilterQueries(fl, OP_AND)
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
	var query string
	filter = ensureComponentVersionFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	filterStr := getComponentVersionFilterString(filter)
	joins := s.getComponentVersionJoins(filter, order)
	cursorFields, err := DecodeCursor(filter.After)
	if err != nil {
		return nil, nil, err
	}

	cursorQuery := CreateCursorQuery("", cursorFields)
	columns := s.getComponentVersionColumns(order)
	order = GetDefaultOrder(order, entity.ComponentVersionId, entity.OrderDirectionAsc)
	orderStr := CreateOrderString(order)

	whereClause := ""
	if filterStr != "" {
		whereClause = fmt.Sprintf("WHERE %s", filterStr)
	}

	if withCursor && cursorQuery != "" {
		// cursor uses aggregated values that need to be used with having
		cursorQuery = fmt.Sprintf("HAVING (%s)", cursorQuery)
	}

	// construct final query
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

	// adding parameters
	var filterParameters []interface{}
	filterParameters = buildQueryParameters(filterParameters, filter.Id)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueId)
	filterParameters = buildQueryParameters(filterParameters, filter.ComponentId)
	filterParameters = buildQueryParameters(filterParameters, filter.Version)
	filterParameters = buildQueryParameters(filterParameters, filter.Tag)
	filterParameters = buildQueryParameters(filterParameters, filter.Repository)
	filterParameters = buildQueryParameters(filterParameters, filter.Organization)
	filterParameters = buildQueryParameters(filterParameters, filter.ComponentCCRN)
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceCCRN)
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceId)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueRepositoryId)
	if withCursor {
		filterParameters = append(filterParameters, GetCursorQueryParameters(filter.First, cursorFields)...)
	}

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetAllComponentVersionIds(filter *entity.ComponentVersionFilter) ([]int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetComponentVersionIds",
	})

	baseQuery := `
		SELECT CV.componentversion_id %s FROM ComponentVersion CV 
		%s
	 	%s GROUP BY CV.componentversion_id ORDER BY %s
    `

	stmt, filterParameters, err := s.buildComponentVersionStatement(baseQuery, filter, false, []entity.Order{}, l)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			l.Warnf("error during closing statement: %s", err)
		}
	}()

	return performIdScan(stmt, filterParameters, l)
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
	defer func() {
		if err := stmt.Close(); err != nil {
			l.Warnf("error during closing statement: %s", err)
		}
	}()

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
	defer func() {
		if err := stmt.Close(); err != nil {
			l.Warnf("error during closing statement: %s", err)
		}
	}()

	return performCountScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) CreateComponentVersion(componentVersion *entity.ComponentVersion) (*entity.ComponentVersion, error) {
	l := logrus.WithFields(logrus.Fields{
		"componentVersion": componentVersion,
		"event":            "database.CreateComponentVersion",
	})

	query := `
		INSERT INTO ComponentVersion (
			componentversion_component_id,
			componentversion_version,
			componentversion_tag,
			componentversion_repository,
			componentversion_organization,
			componentversion_created_by,
			componentversion_updated_by
		) VALUES (
			:componentversion_component_id,
			:componentversion_version,
			:componentversion_tag,
			:componentversion_repository,
			:componentversion_organization,
			:componentversion_created_by,
			:componentversion_updated_by
		)
	`

	componentVersionRow := ComponentVersionRow{}
	componentVersionRow.FromComponentVersion(componentVersion)

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

	updateFields := getComponentVersionUpdateFields(componentVersion)

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
