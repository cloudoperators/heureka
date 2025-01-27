// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type Field struct {
	Name  DbColumnName
	Value any
	Order OrderDirection
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
		case OrderDirectionAsc:
			dir = ">"
		case OrderDirectionDesc:
			dir = "<"
		}
		if i >= len(fields)-1 {
			subQuery = fmt.Sprintf("%s %s %s ? ", subQuery, f.Name, dir)
		} else {
			subQuery = fmt.Sprintf("%s %s = ? AND ", subQuery, f.Name)
		}
	}

	subQuery = fmt.Sprintf("( %s )", subQuery)
	if query != "" {
		subQuery = fmt.Sprintf("%s OR %s", subQuery, query)
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

func WithIssueMatch(order []Order, im IssueMatch) NewCursor {

	return func(cursors *cursors) error {
		cursors.fields = append(cursors.fields, Field{Name: IssueMatchId, Value: im.Id, Order: OrderDirectionAsc})
		cursors.fields = append(cursors.fields, Field{Name: IssueMatchTargetRemediationDate, Value: im.TargetRemediationDate, Order: OrderDirectionAsc})
		cursors.fields = append(cursors.fields, Field{Name: IssueMatchRating, Value: im.Severity.Value, Order: OrderDirectionAsc})

		if im.ComponentInstance != nil {
			cursors.fields = append(cursors.fields, Field{Name: ComponentInstanceCcrn, Value: im.ComponentInstance.CCRN, Order: OrderDirectionAsc})
		}
		if im.Issue != nil {
			cursors.fields = append(cursors.fields, Field{Name: IssuePrimaryName, Value: im.Issue.PrimaryName, Order: OrderDirectionAsc})
		}

		m := CreateOrderMap(order)
		for _, f := range cursors.fields {
			if orderDirection, ok := m[f.Name]; ok {
				f.Order = orderDirection
			}
		}

		return nil
	}
}
