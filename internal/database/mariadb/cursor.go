// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/cloudoperators/heureka/internal/entity"
)

type Field struct {
	Name  entity.OrderByField
	Value any
	Order entity.OrderDirection
}

type cursors struct {
	fields []Field
}

type NewCursor func(cursors *cursors) error

func EncodeCursor(opts ...NewCursor) (string, error) {
	var cursors cursors
	for _, opt := range opts {
		err := opt(&cursors)
		if err != nil {
			return "", err
		}
	}

	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	err := json.NewEncoder(encoder).Encode(cursors.fields)
	if err != nil {
		return "", err
	}
	encoder.Close()
	return buf.String(), nil
}

func DecodeCursor(cursor *string) ([]Field, error) {
	var fields []Field
	if cursor == nil || *cursor == "" {
		return fields, nil
	}
	decoded, err := base64.StdEncoding.DecodeString(*cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 string: %w", err)
	}

	if err := json.Unmarshal(decoded, &fields); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return fields, nil
}

func CreateCursorQuery(query string, fields []Field) string {
	if len(fields) == 0 {
		return query
	}

	subQuery := ""
	for i, f := range fields {
		dir := ">"
		switch f.Order {
		case entity.OrderDirectionAsc:
			dir = ">"
		case entity.OrderDirectionDesc:
			dir = "<"
		}
		if i >= len(fields)-1 {
			subQuery = fmt.Sprintf("%s %s %s ? ", subQuery, ColumnName(f.Name), dir)
		} else {
			subQuery = fmt.Sprintf("%s %s = ? AND ", subQuery, ColumnName(f.Name))
		}
	}

	subQuery = fmt.Sprintf("( %s )", subQuery)
	if query != "" {
		subQuery = fmt.Sprintf("%s OR %s", query, subQuery)
	}

	return CreateCursorQuery(subQuery, fields[:len(fields)-1])
}

func CreateCursorParameters(params []any, fields []Field) []any {
	if len(fields) == 0 {
		return params
	}

	for i := 0; i < len(fields); i++ {
		params = append(params, fields[i].Value)
	}

	return CreateCursorParameters(params, fields[:len(fields)-1])
}

func WithIssueMatch(order []entity.Order, im entity.IssueMatch) NewCursor {

	return func(cursors *cursors) error {
		order = GetDefaultOrder(order, entity.IssueMatchId, entity.OrderDirectionAsc)
		for _, o := range order {
			switch o.By {
			case entity.IssueMatchId:
				cursors.fields = append(cursors.fields, Field{Name: entity.IssueMatchId, Value: im.Id, Order: o.Direction})
			case entity.IssueMatchTargetRemediationDate:
				cursors.fields = append(cursors.fields, Field{Name: entity.IssueMatchTargetRemediationDate, Value: im.TargetRemediationDate, Order: o.Direction})
			case entity.IssueMatchRating:
				cursors.fields = append(cursors.fields, Field{Name: entity.IssueMatchRating, Value: im.Severity.Value, Order: o.Direction})
			case entity.ComponentInstanceCcrn:
				if im.ComponentInstance != nil {
					cursors.fields = append(cursors.fields, Field{Name: entity.ComponentInstanceCcrn, Value: im.ComponentInstance.CCRN, Order: o.Direction})
				}
			case entity.IssuePrimaryName:
				if im.Issue != nil {
					cursors.fields = append(cursors.fields, Field{Name: entity.IssuePrimaryName, Value: im.Issue.PrimaryName, Order: o.Direction})
				}
			default:
				continue
			}
		}

		return nil
	}
}
