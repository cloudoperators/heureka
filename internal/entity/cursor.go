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

func EncodeCursor(order []Order, opts ...NewCursor) (string, error) {
	var cursors cursors
	for _, opt := range opts {
		err := opt(&cursors)
		if err != nil {
			fmt.Println("err")
			return "", err
		}
	}

	m := CreateOrderMap(order)
	for _, f := range cursors.fields {
		if orderDirection, ok := m[f.Name]; ok {
			f.Order = orderDirection
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

func DecodeCursor(cursor string) ([]Field, error) {
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 string: %w", err)
	}

	var fields []Field
	if err := json.Unmarshal(decoded, &fields); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return fields, nil
}

func WithIssueMatch(im IssueMatch) NewCursor {
	return func(cursors *cursors) error {
		cursors.fields = append(cursors.fields, Field{Name: IssueMatchId, Value: im.Id, Order: OrderDirectionAsc})
		// cursors.fields = append(cursors.fields, Field{Name: IssueMatchRating, Value: im.Rating, Order: OrderDirectionAsc})
		return nil
	}
}
